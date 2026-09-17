package cassandra

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/Netcracker/qubership-cassandra-operator/api/v1alpha1"
	"github.com/Netcracker/qubership-cassandra-operator/pkg/impl/utils"
	"github.com/Netcracker/qubership-nosqldb-operator-core/pkg/constants"
	"github.com/Netcracker/qubership-nosqldb-operator-core/pkg/core"
	"github.com/Netcracker/qubership-nosqldb-operator-core/pkg/steps"
	"go.uber.org/zap"
	v13 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type Cassandra struct {
	core.MicroServiceCompound
}

func (r *Cassandra) Validate(ctx core.ExecutionContext) error {
	spec := ctx.Get(constants.ContextSpec).(*v1alpha1.CassandraDeployment)
	if reflect.ValueOf(spec).IsNil() {
		return &core.ExecutionError{Msg: "CassandraService CR spec is not found"}
	}
	return r.DefaultCompound.Validate(ctx)
}

type CassandraBuilder struct {
	core.ExecutableBuilder
}

func (r *CassandraBuilder) Build(ctx core.ExecutionContext) core.Executable {
	spec := ctx.Get(constants.ContextSpec).(*v1alpha1.CassandraDeployment)
	log := ctx.Get(constants.ContextLogger).(*zap.Logger)
	pvcSelector := map[string]string{
		utils.Service: utils.CassandraCluster,
	}

	cassandra := Cassandra{}
	cassandra.ServiceName = utils.Cassandra
	cassandra.CalcDeployType = func(ctx core.ExecutionContext) (core.MicroServiceDeployType, error) {
		request := ctx.Get(constants.ContextRequest).(reconcile.Request)

		helperImpl := ctx.Get(utils.KubernetesHelperImpl).(core.KubernetesHelper)

		pvcList := &v13.PersistentVolumeClaimList{}
		err := helperImpl.ListRuntimeObjectsByLabels(pvcList, request.Namespace, pvcSelector)
		var result core.MicroServiceDeployType
		if err != nil {
			result = core.Empty
		} else if len(pvcList.Items) == 0 {
			result = core.CleanDeploy
		} else {
			result = core.Update
		}

		log.Debug(fmt.Sprintf("%s deploy mode is used for %s service", result, utils.Cassandra))

		return result, err
	}

	//cassandra.AddStep(steps.NewMaintenanceConsulServiceStep(
	//	utils.Cassandra,
	//	CassandraConsulServiceRegistrationCast,
	//	true,
	//	"Deploying"))

	for index, dc := range utils.FilterDC(spec.Spec.Cassandra.DeploymentSchema.DataCenters, func(dc *v1alpha1.DataCenter) bool { return dc.Deploy }) {
		pvcContextFormat := fmt.Sprintf(utils.CassandraDCPvcNameFormat, index)
		nodesContext := fmt.Sprintf(utils.PVNodesFormat, index)
		replicas := dc.GetActiveReplicas()
		for storageIndex, storage := range dc.Storage {
			// PVC are stored in the context per storage
			pvcContext := fmt.Sprintf("%s-%v", pvcContextFormat, storageIndex)

			for _, replica := range replicas {
				pvcNameFormat := pvcContextFormat + "-%v"
				if storageIndex != 0 {
					// For any additional storages storage index in included
					pvcNameFormat = pvcNameFormat + fmt.Sprintf("-%v", storageIndex)
					// As the result the format for PVC is looks like that:
					// cassandra-data-dc%dcIndex%-%replicaIndex%-%storageIndexMoreThanZero%
				}
				pvcStep := &steps.CreatePVCStep{
					Storage:           storage,
					NameFormat:        pvcNameFormat,
					LabelSelector:     pvcSelector,
					ContextVarToStore: pvcContext,
					PVCCount: func(ctx core.ExecutionContext) int {
						return 1
					},
					WaitTimeout:  spec.Spec.WaitTimeout,
					Owner:        nil,
					WaitPVCBound: storage.WaitPVCBound,
					StartIndex:   replica,
				}

				if spec.Spec.DeletePVConUninstall {
					pvcStep.Owner = spec
				}

				cassandra.AddStep(pvcStep)
			}

			//Do it only for main storage. The rest should line with nodes from main storage
			if storageIndex == 0 {
				cassandra.AddStep(&steps.StoreNodesStep{
					Storage:           storage,
					ContextVarToStore: nodesContext,
				})
			}

			var tolerations []v13.Toleration
			if spec.Spec.Policies != nil {
				tolerations = spec.Spec.Policies.Tolerations
			}

			if spec.Spec.Recycler.Install {
				// Perform recycling for each pvc in each storage with the nodes from main storage
				cassandra.AddStep(&steps.PVRecyclerStep{
					DockerImage:        spec.Spec.Cassandra.DockerImage,
					Volumes:            storage.Volumes,
					Tolerations:        tolerations,
					PVCContextVar:      pvcContext,
					PVNodesContextVar:  nodesContext,
					WaitTimeout:        spec.Spec.WaitTimeout,
					PodSecurityContext: spec.Spec.PodSecurityContext,
					Resources:          spec.Spec.Recycler.Resources,
					Owner:              spec,
				})
			}
		}

		// When commitlog archiving is enabled with a dedicated PVC, create one
		// archive PVC per replica so archived commit logs never share space with
		// the main data volume.
		commitlogArchiving := spec.Spec.Cassandra.CommitlogArchiving
		if commitlogArchiving.Enabled && commitlogArchiving.Storage != nil {
			archivePvcStorage := commitlogArchiving.Storage

			// Ensure the fixed mount settings are applied regardless of what is
			// specified in the storage field (mountSettings is not user-configurable
			// for the archive PVC — the path is fixed by design).
			archiveMountSettings := &v13.VolumeMount{
				Name:      utils.CassandraCommitlogArchivesMountName,
				MountPath: utils.CassandraCommitlogArchivesMountPath,
			}
			archivePvcStorage.MountSettings = archiveMountSettings

			archivePvcContext := fmt.Sprintf(utils.CassandraCommitlogArchivesPvcContext, index)
			archivePvcNameFormat := fmt.Sprintf(utils.CassandraDCCommitlogArchivesPvcNameFormat, index) + "-%v"

			for _, replica := range replicas {
				archivePvcStep := &steps.CreatePVCStep{
					Storage:           archivePvcStorage,
					NameFormat:        archivePvcNameFormat,
					LabelSelector:     pvcSelector,
					ContextVarToStore: archivePvcContext,
					PVCCount: func(ctx core.ExecutionContext) int {
						return 1
					},
					WaitTimeout:  spec.Spec.WaitTimeout,
					Owner:        nil,
					WaitPVCBound: archivePvcStorage.WaitPVCBound,
					StartIndex:   replica,
				}

				if spec.Spec.DeletePVConUninstall {
					archivePvcStep.Owner = spec
				}

				cassandra.AddStep(archivePvcStep)
			}
		}

	}

	cassandra.AddStep(&CassandraServicesStep{})
	cassandra.AddStep(&CassandraLoadbalancerService{})

	cassandra.AddStep(&CollectCassandraPVCsStep{})
	cassandra.AddStep(&steps.WaitForPVCExpansionStep{
		WaitTimeout: spec.Spec.WaitTimeout,
		PVCNamesVar: utils.CassandraAllPVCsContext,
		OnNeedsRestart: func(ctx core.ExecutionContext) error {
			return restartCassandraStatefulSets(ctx, spec)
		},
	})

	cassandra.AddStep(&CassandraStatefulSetStep{})

	cassandra.AddStep(&CreateSuperUser{
		Username: spec.Spec.User,
		Password: func() string { return ctx.Get(utils.ContextPasswordKey).(string) },
	})

	cassandra.AddStep(&UpdateCassandraCredentials{})

	cassandra.AddStep(&RemoveNodes{})

	cassandra.AddStep(&CleanupNodes{})

	cassandra.AddStep(&UpdateSystemKeyspacesTopology{})

	cassandra.AddStep(&CassandraReaper{})

	cassandra.AddStep(&NodetoolRebuild{})

	cassandra.AddStep(&DropCassandraDefaultUser{})

	//cassandra.AddStep(steps.NewMaintenanceConsulServiceStep(
	//	utils.Cassandra,
	//	CassandraConsulServiceRegistrationCast,
	//	false))

	return &cassandra
}

func (r *Cassandra) Condition(ctx core.ExecutionContext) (bool, error) {
	spec := ctx.Get(constants.ContextSpec).(*v1alpha1.CassandraDeployment)
	microServiceCheck, microserviceCheckErr := core.CheckSpecChange(ctx, spec.Spec.Cassandra, utils.Cassandra)
	commonCheck := ctx.Get(constants.IsAnyCommonParameterChanged).(bool)

	if microserviceCheckErr != nil {
		return microServiceCheck, microserviceCheckErr
	} else {
		return microServiceCheck || commonCheck, nil
	}
}

// restartCassandraStatefulSets cycles each StatefulSet one at a time to keep the cluster
// available during PVC expansion. For each pod: scale down, wait 60s for Cinder to detach
// and complete the block resize, scale back up, then wait for CQL login before moving to
// the next pod.
func restartCassandraStatefulSets(ctx core.ExecutionContext, spec *v1alpha1.CassandraDeployment) error {
	helperImpl := ctx.Get(utils.KubernetesHelperImpl).(core.KubernetesHelper)
	cassandraHelperImpl := ctx.Get(utils.CassandraHelperImpl).(utils.CassandraUtils)
	request := ctx.Get(constants.ContextRequest).(reconcile.Request)
	log := ctx.Get(constants.ContextLogger).(*zap.Logger)

	username := spec.Spec.User
	password := ctx.Get(utils.ContextPasswordKey).(string)

	dcReplicas := utils.FilterDC(spec.Spec.Cassandra.DeploymentSchema.DataCenters, func(dc *v1alpha1.DataCenter) bool { return dc.Deploy })

	for dcIndex, dc := range dcReplicas {
		for _, replicaIndex := range dc.GetActiveReplicas() {
			ssName := fmt.Sprintf(utils.CassandraReplicaNameFormat, utils.CalcReplicaIndex(dcReplicas, dcIndex, replicaIndex))

			log.Info(fmt.Sprintf("Scaling down %s for PVC filesystem resize", ssName))
			if err := helperImpl.ScaleStatefulSetByName(ssName, request.Namespace, 0, spec.Spec.WaitTimeout); err != nil {
				return err
			}

			// Collect PVC names for this replica to verify capacity after scale-up.
			pvcContextFormat := fmt.Sprintf(utils.CassandraDCPvcNameFormat, dcIndex)
			var replicaPVCs []string
			for storageIndex := range dc.Storage {
				if storageIndex == 0 {
					replicaPVCs = append(replicaPVCs, fmt.Sprintf("%s-%v", pvcContextFormat, replicaIndex))
				} else {
					replicaPVCs = append(replicaPVCs, fmt.Sprintf("%s-%v-%v", pvcContextFormat, replicaIndex, storageIndex))
				}
			}
			commitlogArchiving := spec.Spec.Cassandra.CommitlogArchiving
			if commitlogArchiving.Enabled && commitlogArchiving.Storage != nil {
				replicaPVCs = append(replicaPVCs, fmt.Sprintf(utils.CassandraDCCommitlogArchivesPvcNameFormat+"-%v", dcIndex, replicaIndex))
			}

			log.Info(fmt.Sprintf("Scaling up %s", ssName))
			if err := helperImpl.ScaleStatefulSetByName(ssName, request.Namespace, 1, spec.Spec.WaitTimeout); err != nil {
				return err
			}

			// Wait until kubelet completes NodeExpandVolume for each PVC, confirmed by
			// Status.Capacity reaching the requested size. FileSystemResizePending=true only
			// means the Cinder control-plane accepted the request — the block device on the
			// node is not resized until the pod mounts the volume.
			k8sClient := ctx.Get(constants.ContextClient).(client.Client)
			for _, pvcName := range replicaPVCs {
				log.Info(fmt.Sprintf("Waiting for PVC %s capacity to reflect new size after %s restart", pvcName, ssName))
				if err := wait.PollImmediate(5*time.Second, time.Duration(spec.Spec.WaitTimeout)*time.Second,
					func() (bool, error) {
						pvc := &v13.PersistentVolumeClaim{}
						if err := k8sClient.Get(context.TODO(), types.NamespacedName{Name: pvcName, Namespace: request.Namespace}, pvc); err != nil {
							return false, err
						}
						requested := pvc.Spec.Resources.Requests[v13.ResourceStorage]
						capacity := pvc.Status.Capacity[v13.ResourceStorage]
						done := capacity.Cmp(requested) >= 0
						if !done {
							log.Info(fmt.Sprintf("PVC %s capacity %s < requested %s, retrying", pvcName, capacity.String(), requested.String()))
						}
						return done, nil
					},
				); err != nil {
					return fmt.Errorf("PVC %s did not reach requested capacity after restarting %s: %w", pvcName, ssName, err)
				}
			}

			// Wait for this node to rejoin the cluster before cycling the next one.
			log.Info(fmt.Sprintf("Waiting for Cassandra to become healthy after %s restart", ssName))
			if err := wait.PollImmediate(10*time.Second, time.Duration(spec.Spec.WaitTimeout)*time.Second,
				func() (bool, error) {
					if !cassandraHelperImpl.CheckLogin(ctx, username, password) {
						log.Info("Cassandra not yet accepting connections, retrying")
						return false, nil
					}
					return true, nil
				},
			); err != nil {
				return fmt.Errorf("cassandra not healthy after restarting %s: %w", ssName, err)
			}
		}
	}
	return nil
}
