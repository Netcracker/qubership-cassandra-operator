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

package cassandra

import (
	"fmt"

	"github.com/Netcracker/qubership-cassandra-operator/api/v1alpha1"
	"github.com/Netcracker/qubership-cassandra-operator/pkg/impl/utils"
	"github.com/Netcracker/qubership-nosqldb-operator-core/pkg/constants"
	"github.com/Netcracker/qubership-nosqldb-operator-core/pkg/core"
	"go.uber.org/zap"
)

type NodetoolRebuild struct {
	core.DefaultExecutable
}

func (r *NodetoolRebuild) Execute(ctx core.ExecutionContext) error {
	log := ctx.Get(constants.ContextLogger).(*zap.Logger)
	spec := ctx.Get(constants.ContextSpec).(*v1alpha1.CassandraDeployment)
	cassandraHelperImpl := ctx.Get(utils.CassandraHelperImpl).(utils.CassandraUtils)
	existingDC := utils.NewStream(spec.Spec.Cassandra.DeploymentSchema.DataCenters).FindFirst(func(dc interface{}) bool {
		return !dc.(*v1alpha1.DataCenter).Deploy
	}).(*v1alpha1.DataCenter)

	log.Debug(fmt.Sprintf("All replicas are deployed, running nodetool rebuild -- %s", existingDC.Name))

	list, err := cassandraHelperImpl.GetAllCassandraPods(ctx)
	core.PanicError(err, log.Error, "Cassandra pods listing failed")

	if len(list.Items) == 0 {
		return &core.ExecutionError{Msg: "Cassandra pods not found"}
	}
	for _, cassandraPod := range list.Items {
		_, err := cassandraHelperImpl.RunSshOnPod(&cassandraPod, ctx,
			fmt.Sprintf("nodetool rebuild -- %s 2> /dev/null", existingDC.Name))
		core.PanicError(err, log.Error, "Could not rebuild node")
	}

	return nil
}

func (r *NodetoolRebuild) Condition(ctx core.ExecutionContext) (bool, error) {
	cassandraHelperImpl := ctx.Get(utils.CassandraHelperImpl).(utils.CassandraUtils)
	return core.GetCurrentDeployType(ctx) == core.CleanDeploy && cassandraHelperImpl.IsAllDCsDeployed(ctx), nil
}
