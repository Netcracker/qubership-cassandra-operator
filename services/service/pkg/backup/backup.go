package backup

import (
	"context"
	"fmt"
	"time"

	v1 "github.com/Netcracker/qubership-cassandra-supplementary/api/v1alpha1"
	"github.com/Netcracker/qubership-cassandra-supplementary/pkg/utils"
	"github.com/Netcracker/qubership-nosqldb-operator-core/pkg/constants"
	"github.com/Netcracker/qubership-nosqldb-operator-core/pkg/core"
	"github.com/Netcracker/qubership-nosqldb-operator-core/pkg/steps"
	"go.uber.org/zap"
	v12 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type CassandraBackup struct {
	core.MicroServiceCompound
}

type BackupBuilder struct {
	core.ExecutableBuilder
}

func (r *BackupBuilder) Build(ctx core.ExecutionContext) core.Executable {
	spec := ctx.Get(constants.ContextSpec).(*v1.CassandraSupplService)

	backupSpec := spec.Spec.Backup
	storage := backupSpec.Storage

	pvcSelector := map[string]string{
		utils.Name: utils.BackupDaemon,
	}

	backup := CassandraBackup{}
	backup.ServiceName = utils.Backup
	backup.CalcDeployType = func(ctx core.ExecutionContext) (core.MicroServiceDeployType, error) {
		request := ctx.Get(constants.ContextRequest).(reconcile.Request)
		log := ctx.Get(constants.ContextLogger).(*zap.Logger)
		helperImpl := ctx.Get(utils.KubernetesHelperImpl).(core.KubernetesHelper)

		pvcList := &v12.PersistentVolumeClaimList{}
		err := helperImpl.ListRuntimeObjectsByLabels(pvcList, request.Namespace, pvcSelector)
		var result core.MicroServiceDeployType
		if err != nil {
			result = core.Empty
		} else if len(pvcList.Items) == 0 {
			result = core.CleanDeploy
		} else {
			result = core.Update
		}

		if err == nil {
			log.Debug(fmt.Sprintf("%s deploy mode is used for %s service", result, utils.Backup))
		}

		return result, err
	}

	pvcContext := fmt.Sprintf(utils.BackupPvcName, 0)
	nodesContext := fmt.Sprintf(utils.PVNodesFormat, 0)

	if !spec.Spec.Backup.Storage.EmptyDir {
		pvcStep := &steps.CreatePVCStep{
			Storage:           storage,
			NameFormat:        utils.BackupPvcName,
			LabelSelector:     pvcSelector,
			ContextVarToStore: pvcContext,
			PVCCount: func(ctx core.ExecutionContext) int {
				return 1
			},
			WaitTimeout:  spec.Spec.WaitTimeout,
			Owner:        nil,
			WaitPVCBound: spec.Spec.Backup.Storage.WaitPVCBound,
		}

		if spec.Spec.DeletePVConUninstall {
			pvcStep.Owner = spec
		}

		backup.AddStep(pvcStep)
		backup.AddStep(&steps.StoreNodesStep{
			Storage:           storage,
			ContextVarToStore: nodesContext,
		})
		backupWaitSeconds := spec.Spec.WaitTimeout
		backup.AddStep(&steps.WaitForPVCExpansionStep{
			WaitTimeout: backupWaitSeconds,
			PVCNamesVar: pvcContext,
			OnNeedsRestart: func(ctx core.ExecutionContext) error {
				request := ctx.Get(constants.ContextRequest).(reconcile.Request)
				helperImpl := ctx.Get(utils.KubernetesHelperImpl).(core.KubernetesHelper)
				k8sClient := ctx.Get(constants.ContextClient).(client.Client)
				log := ctx.Get(constants.ContextLogger).(*zap.Logger)

				log.Info(fmt.Sprintf("Scaling deployment %s down for volume resize", utils.BackupDaemon))
				if err := helperImpl.ScaleDeploymentByLabels(
					map[string]string{utils.Name: utils.BackupDaemon},
					request.Namespace, 0, backupWaitSeconds,
				); err != nil {
					return fmt.Errorf("scaling down %s: %w", utils.BackupDaemon, err)
				}

				// Wait for ControllerExpandVolume to finish before mounting the volume.
				// NodeExpandVolume (which runs when the pod mounts the volume) fails with
				// "current volume size is less than expected" when the storage backend has
				// not yet resized the block device. Waiting here ensures the backend has
				// committed the new size before the pod attaches.
				pvcName := fmt.Sprintf(utils.BackupPvcName, 0)
				log.Info(fmt.Sprintf("Waiting for PVC %s expansion to complete before scaling up %s", pvcName, utils.BackupDaemon))
				if err := wait.PollImmediate(5*time.Second, time.Duration(backupWaitSeconds)*time.Second,
					func() (bool, error) {
						pvc := &v12.PersistentVolumeClaim{}
						if err := k8sClient.Get(context.TODO(), types.NamespacedName{Name: pvcName, Namespace: request.Namespace}, pvc); err != nil {
							return false, err
						}
						requested := pvc.Spec.Resources.Requests[v12.ResourceStorage]
						capacity := pvc.Status.Capacity[v12.ResourceStorage]
						if capacity.Cmp(requested) >= 0 {
							return true, nil
						}
						log.Info(fmt.Sprintf("PVC %s capacity %s < requested %s, waiting for ControllerExpand", pvcName, capacity.String(), requested.String()))
						return false, nil
					},
				); err != nil {
					return fmt.Errorf("PVC %s did not reach requested capacity before scaling up %s: %w", pvcName, utils.BackupDaemon, err)
				}

				log.Info(fmt.Sprintf("Scaling deployment %s up after PVC expansion", utils.BackupDaemon))
				if err := helperImpl.ScaleDeploymentByLabels(
					map[string]string{utils.Name: utils.BackupDaemon},
					request.Namespace, 1, backupWaitSeconds,
				); err != nil {
					return fmt.Errorf("scaling up %s: %w", utils.BackupDaemon, err)
				}
				return nil
			},
		})
	}

	backup.AddStep(&BackupService{})

	if !spec.Spec.AWSKeyspaces.Install {
		backup.AddStep(&BackupSSHKeyStep{})
	}

	backup.AddStep(&LegacyBackupDeployment{})

	return &backup
}

func (r *CassandraBackup) Condition(ctx core.ExecutionContext) (bool, error) {
	spec := ctx.Get(constants.ContextSpec).(*v1.CassandraSupplService)
	microServiceCheck, microserviceCheckErr := core.CheckSpecChange(ctx, spec.Spec.Backup, utils.BackupDaemon)
	commonCheck := ctx.Get(constants.IsAnyCommonParameterChanged).(bool)

	if microserviceCheckErr != nil {
		return microServiceCheck, microserviceCheckErr
	} else {
		return microServiceCheck || commonCheck, nil
	}
}
