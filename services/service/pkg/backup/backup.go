package backup

import (
	"fmt"
	"time"

	v1 "github.com/Netcracker/qubership-cassandra-supplementary/api/v1alpha1"
	"github.com/Netcracker/qubership-cassandra-supplementary/pkg/utils"
	"github.com/Netcracker/qubership-nosqldb-operator-core/pkg/constants"
	"github.com/Netcracker/qubership-nosqldb-operator-core/pkg/core"
	"github.com/Netcracker/qubership-nosqldb-operator-core/pkg/steps"
	"go.uber.org/zap"
	v12 "k8s.io/api/core/v1"
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
				log := ctx.Get(constants.ContextLogger).(*zap.Logger)

				// Scale down to 0.
				log.Info(fmt.Sprintf("Scaling deployment %s down for volume resize", utils.BackupDaemon))
				if err := helperImpl.ScaleDeploymentByLabels(
					map[string]string{utils.Name: utils.BackupDaemon},
					request.Namespace, 0, backupWaitSeconds,
				); err != nil {
					return fmt.Errorf("scaling down %s: %w", utils.BackupDaemon, err)
				}

				// Wait for Cinder to detach the volume and complete block-level resize.
				log.Info(fmt.Sprintf("Deployment %s down, waiting 60s for volume detach and block resize", utils.BackupDaemon))
				time.Sleep(60 * time.Second)

				// Scale back up with retry.
				const maxAttempts = 5
				for attempt := 1; attempt <= maxAttempts; attempt++ {
					delay := time.Duration(10*(1<<uint(attempt-1))) * time.Second
					log.Info(fmt.Sprintf("Attempt %d/%d: waiting %s before starting %s", attempt, maxAttempts, delay, utils.BackupDaemon))
					time.Sleep(delay)

					if err := helperImpl.ScaleDeploymentByLabels(
						map[string]string{utils.Name: utils.BackupDaemon},
						request.Namespace, 1, backupWaitSeconds,
					); err == nil {
						log.Info(fmt.Sprintf("Deployment %s started successfully on attempt %d", utils.BackupDaemon, attempt))
						return nil
					} else if attempt == maxAttempts {
						return fmt.Errorf("deployment %s failed to start after %d attempts: %w", utils.BackupDaemon, maxAttempts, err)
					}
					log.Warn(fmt.Sprintf("Deployment %s not healthy on attempt %d, retrying", utils.BackupDaemon, attempt))

					// Scale back down before next attempt.
					_ = helperImpl.ScaleDeploymentByLabels(
						map[string]string{utils.Name: utils.BackupDaemon},
						request.Namespace, 0, backupWaitSeconds,
					)
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
