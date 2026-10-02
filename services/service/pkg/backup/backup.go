// Copyright 2024-2025 NetCracker Technology Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
		backup.AddStep(&checkPVCFilesystemResizePendingStep{pvcName: fmt.Sprintf(utils.BackupPvcName, 0)})
		backupWaitSeconds := spec.Spec.WaitTimeout
		backup.AddStep(&steps.WaitForPVCExpansionStep{
			WaitTimeout: backupWaitSeconds,
			PVCNamesVar: pvcContext,
			OnNeedsRestart: func(ctx core.ExecutionContext) error {
				request := ctx.Get(constants.ContextRequest).(reconcile.Request)
				helperImpl := ctx.Get(utils.KubernetesHelperImpl).(core.KubernetesHelper)
				log := ctx.Get(constants.ContextLogger).(*zap.Logger)

				log.Info(fmt.Sprintf("Scaling deployment %s down for PVC filesystem resize", utils.BackupDaemon))
				if err := helperImpl.ScaleDeploymentByLabels(
					map[string]string{utils.Name: utils.BackupDaemon},
					request.Namespace, 0, backupWaitSeconds,
				); err != nil {
					return fmt.Errorf("scaling down %s: %w", utils.BackupDaemon, err)
				}
				const maxAttempts = 5
				for attempt := 1; attempt <= maxAttempts; attempt++ {
					delay := time.Duration(60*(1<<uint(attempt-1))) * time.Second
					log.Info(fmt.Sprintf("Attempt %d/%d: waiting %s before scaling up %s", attempt, maxAttempts, delay, utils.BackupDaemon))
					time.Sleep(delay)

					log.Info(fmt.Sprintf("Scaling deployment %s up (attempt %d/%d)", utils.BackupDaemon, attempt, maxAttempts))
					if err := helperImpl.ScaleDeploymentByLabels(
						map[string]string{utils.Name: utils.BackupDaemon},
						request.Namespace, 1, backupWaitSeconds,
					); err == nil {
						log.Info(fmt.Sprintf("Deployment %s started successfully on attempt %d", utils.BackupDaemon, attempt))
						return nil
					} else if attempt == maxAttempts {
						return fmt.Errorf("deployment %s failed to start after %d attempts: %w", utils.BackupDaemon, maxAttempts, err)
					}

					log.Info(fmt.Sprintf("Deployment %s not healthy on attempt %d, scaling back down for retry", utils.BackupDaemon, attempt))
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

// checkPVCFilesystemResizePendingStep sets PVCResizeNeeded=true in the context
// when the PVC still has FileSystemResizePending=True. This ensures
// WaitForPVCExpansionStep is not skipped in reconciles where the PVC spec was
// already updated (PVCResizeNeeded was not set by CreatePVCStep).
type checkPVCFilesystemResizePendingStep struct {
	core.DefaultExecutable
	pvcName string
}

func (r *checkPVCFilesystemResizePendingStep) Execute(ctx core.ExecutionContext) error {
	request := ctx.Get(constants.ContextRequest).(reconcile.Request)
	k8sClient := ctx.Get(constants.ContextClient).(client.Client)

	pvc := &v12.PersistentVolumeClaim{}
	if err := k8sClient.Get(context.TODO(), types.NamespacedName{Name: r.pvcName, Namespace: request.Namespace}, pvc); err != nil {
		return fmt.Errorf("getting PVC %s: %w", r.pvcName, err)
	}
	for _, cond := range pvc.Status.Conditions {
		if cond.Type == v12.PersistentVolumeClaimFileSystemResizePending &&
			cond.Status == v12.ConditionTrue {
			ctx.Set(constants.PVCResizeNeeded, true)
			break
		}
	}
	return nil
}

func (r *checkPVCFilesystemResizePendingStep) Condition(ctx core.ExecutionContext) (bool, error) {
	return true, nil
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
