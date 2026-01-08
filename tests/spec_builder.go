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

package tests

import (
	"github.com/Netcracker/qubership-cassandra-operator/api/v1alpha1"
	mTypes "github.com/Netcracker/qubership-nosqldb-operator-core/pkg/types"
	v1core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func GenerateDefaultCassandra(cassandraDCs []*v1alpha1.DataCenter) *v1alpha1.CassandraDeployment {
	GiQuantity, _ := resource.ParseQuantity("5Gi")
	var fsGroup int64 = 999
	var tolerationSeconds int64 = 20

	rr := &v1core.ResourceRequirements{
		Limits: v1core.ResourceList{
			v1core.ResourceMemory: GiQuantity,
		},
		Requests: nil,
	}

	return &v1alpha1.CassandraDeployment{
		Spec: v1alpha1.CassandraSpec{
			WaitTimeout: 100,
			Recycler: mTypes.Recycler{
				Resources: rr,
			},
			ServiceAccountName: "cassandra-operator",
			Policies: &v1alpha1.Policies{
				Tolerations: []v1core.Toleration{
					{
						Key:               "key1",
						Value:             "value1",
						Operator:          v1core.TolerationOpEqual,
						Effect:            v1core.TaintEffectNoSchedule,
						TolerationSeconds: &tolerationSeconds,
					},
					{
						Key:               "key2",
						Value:             "value2",
						Operator:          v1core.TolerationOpEqual,
						Effect:            v1core.TaintEffectNoExecute,
						TolerationSeconds: &tolerationSeconds,
					},
				},
			},
			PodSecurityContext: &v1core.PodSecurityContext{
				FSGroup: &fsGroup,
			},
			Cassandra: v1alpha1.Cassandra{
				Install:          true,
				User:             "admin",
				SecretName:       "cassandra-secret",
				Resources:        rr,
				DeploymentSchema: &v1alpha1.DeploymentSchema{DataCenters: cassandraDCs},
			},
			VaultRegistration: mTypes.VaultRegistration{
				InitContainerResources: rr,
			},
		},
	}
}
