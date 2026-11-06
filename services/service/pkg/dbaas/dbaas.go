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

package dbaas

import (
	v1 "github.com/Netcracker/qubership-cassandra-supplementary/api/v1alpha1"
	"github.com/Netcracker/qubership-cassandra-supplementary/pkg/utils"
	"github.com/Netcracker/qubership-nosqldb-operator-core/pkg/constants"
	"github.com/Netcracker/qubership-nosqldb-operator-core/pkg/core"
)

type DbaasCompound struct {
	core.MicroServiceCompound
}

type DbaasBuilder struct {
	core.ExecutableBuilder
}

func (r *DbaasBuilder) Build(ctx core.ExecutionContext) core.Executable {
	dbaas := DbaasCompound{}
	dbaas.ServiceName = utils.Dbaas
	dbaas.CalcDeployType = func(ctx core.ExecutionContext) (deployType core.MicroServiceDeployType, err error) {
		return core.CleanDeploy, nil
	}

	dbaas.AddStep(&DbaasService{})
	dbaas.AddStep(&DbaasDeployment{})

	return &dbaas
}

func (r *DbaasCompound) Condition(ctx core.ExecutionContext) (bool, error) {
	spec := ctx.Get(constants.ContextSpec).(*v1.CassandraSupplService)
	microServiceCheck, microserviceCheckErr := core.CheckSpecChange(ctx, spec.Spec.Dbaas, utils.DbaasName)
	commonCheck := ctx.Get(constants.IsAnyCommonParameterChanged).(bool)

	if microserviceCheckErr != nil {
		return microServiceCheck, microserviceCheckErr
	} else {
		return microServiceCheck || commonCheck, nil
	}
}
