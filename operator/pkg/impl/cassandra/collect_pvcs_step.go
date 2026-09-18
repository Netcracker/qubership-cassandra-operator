package cassandra

import (
	"fmt"

	"github.com/Netcracker/qubership-cassandra-operator/api/v1alpha1"
	"github.com/Netcracker/qubership-cassandra-operator/pkg/impl/utils"
	"github.com/Netcracker/qubership-nosqldb-operator-core/pkg/constants"
	"github.com/Netcracker/qubership-nosqldb-operator-core/pkg/core"
)

// CollectCassandraPVCsStep gathers all per-DC PVC names created by CreatePVCStep calls
// into a single slice stored under CassandraAllPVCsContext so WaitForPVCExpansionStep
// can wait on all of them at once.
type CollectCassandraPVCsStep struct {
	core.DefaultExecutable
}

func (r *CollectCassandraPVCsStep) Execute(ctx core.ExecutionContext) error {
	spec := ctx.Get(constants.ContextSpec).(*v1alpha1.CassandraDeployment)

	dcReplicas := utils.FilterDC(spec.Spec.Cassandra.DeploymentSchema.DataCenters, func(dc *v1alpha1.DataCenter) bool { return dc.Deploy })

	var all []string
	for dcIndex, dc := range dcReplicas {
		pvcContextFormat := fmt.Sprintf(utils.CassandraDCPvcNameFormat, dcIndex)
		for storageIndex := range dc.Storage {
			pvcContext := fmt.Sprintf("%s-%v", pvcContextFormat, storageIndex)
			if names, ok := ctx.Get(pvcContext).([]string); ok {
				all = append(all, names...)
			}
		}

		commitlogArchiving := spec.Spec.Cassandra.CommitlogArchiving
		if commitlogArchiving.Enabled && commitlogArchiving.Storage != nil {
			archivePvcContext := fmt.Sprintf(utils.CassandraCommitlogArchivesPvcContext, dcIndex)
			if names, ok := ctx.Get(archivePvcContext).([]string); ok {
				all = append(all, names...)
			}
		}
	}

	ctx.Set(utils.CassandraAllPVCsContext, all)
	return nil
}

func (r *CollectCassandraPVCsStep) Condition(ctx core.ExecutionContext) (bool, error) {
	return true, nil
}
