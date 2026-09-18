package backup

import (
	"context"
	"fmt"
	"time"

	"github.com/Netcracker/qubership-cassandra-supplementary/pkg/utils"
	"github.com/Netcracker/qubership-nosqldb-operator-core/pkg/constants"
	"github.com/Netcracker/qubership-nosqldb-operator-core/pkg/core"
	v12 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// BackupPVCFilesystemResizeStep checks whether the backup PVC has a pending
// filesystem resize (FileSystemResizePending=True) and, if so, bounces the
// backup-daemon deployment so kubelet can run NodeExpandVolume on mount.
// This step runs unconditionally so it catches cases where the PVC spec was
// already updated in a prior reconcile (PVCResizeNeeded=false) but the pod
// has not yet been restarted to complete the node-side resize.
type BackupPVCFilesystemResizeStep struct {
	core.DefaultExecutable
	WaitSeconds int
}

func (r *BackupPVCFilesystemResizeStep) Execute(ctx core.ExecutionContext) error {
	request := ctx.Get(constants.ContextRequest).(reconcile.Request)
	helperImpl := ctx.Get(utils.KubernetesHelperImpl).(core.KubernetesHelper)
	k8sClient := ctx.Get(constants.ContextClient).(client.Client)
	log := ctx.Get(constants.ContextLogger).(interface {
		Info(msg string, fields ...interface{})
	})

	pvcName := fmt.Sprintf(utils.BackupPvcName, 0)

	pvc := &v12.PersistentVolumeClaim{}
	if err := k8sClient.Get(context.TODO(), types.NamespacedName{Name: pvcName, Namespace: request.Namespace}, pvc); err != nil {
		return fmt.Errorf("getting PVC %s: %w", pvcName, err)
	}

	needsRestart := false
	for _, cond := range pvc.Status.Conditions {
		if cond.Type == v12.PersistentVolumeClaimFileSystemResizePending &&
			cond.Status == v12.ConditionTrue {
			needsRestart = true
			break
		}
	}

	if !needsRestart {
		return nil
	}

	log.Info(fmt.Sprintf("PVC %s has FileSystemResizePending — bouncing %s to complete node-side resize", pvcName, utils.BackupDaemon))

	if err := helperImpl.ScaleDeploymentByLabels(
		map[string]string{utils.Name: utils.BackupDaemon},
		request.Namespace, 0, r.WaitSeconds,
	); err != nil {
		return fmt.Errorf("scaling down %s: %w", utils.BackupDaemon, err)
	}

	// Scale back up with exponential-backoff retry. Cinder may not have
	// finished the block-level resize by the time the pod first attaches;
	// if it does not become ready within 2 minutes we scale back down and
	// try again with a longer wait.
	const maxAttempts = 5
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		delay := time.Duration(10*(1<<uint(attempt-1))) * time.Second
		log.Info(fmt.Sprintf("Attempt %d/%d: waiting %s before scaling up %s", attempt, maxAttempts, delay, utils.BackupDaemon))
		time.Sleep(delay)

		log.Info(fmt.Sprintf("Scaling deployment %s up (attempt %d/%d)", utils.BackupDaemon, attempt, maxAttempts))
		if err := helperImpl.ScaleDeploymentByLabels(
			map[string]string{utils.Name: utils.BackupDaemon},
			request.Namespace, 1, 120,
		); err == nil {
			break
		} else if attempt == maxAttempts {
			return fmt.Errorf("deployment %s failed to start after %d attempts: %w", utils.BackupDaemon, maxAttempts, err)
		}

		log.Info(fmt.Sprintf("Deployment %s not healthy on attempt %d, scaling back down for retry", utils.BackupDaemon, attempt))
		_ = helperImpl.ScaleDeploymentByLabels(
			map[string]string{utils.Name: utils.BackupDaemon},
			request.Namespace, 0, r.WaitSeconds,
		)
	}

	// Wait until FileSystemResizePending clears, confirming NodeExpandVolume succeeded.
	log.Info(fmt.Sprintf("Waiting for PVC %s filesystem resize to complete", pvcName))
	return wait.PollUntilContextTimeout(context.Background(), 5*time.Second, time.Duration(r.WaitSeconds)*time.Second, true,
		func(ctx context.Context) (bool, error) {
			updated := &v12.PersistentVolumeClaim{}
			if err := k8sClient.Get(ctx, types.NamespacedName{Name: pvcName, Namespace: request.Namespace}, updated); err != nil {
				return false, err
			}
			for _, cond := range updated.Status.Conditions {
				if cond.Type == v12.PersistentVolumeClaimFileSystemResizePending &&
					cond.Status == v12.ConditionTrue {
					log.Info(fmt.Sprintf("PVC %s filesystem resize still pending", pvcName))
					return false, nil
				}
			}
			return true, nil
		},
	)
}

func (r *BackupPVCFilesystemResizeStep) Condition(ctx core.ExecutionContext) (bool, error) {
	return true, nil
}
