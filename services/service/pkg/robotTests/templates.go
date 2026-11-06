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

package robotTests

import (
	"github.com/Netcracker/qubership-cassandra-supplementary/pkg/utils"
	v12 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func RobotTemplate(namespace string,
	image string,
	resources v1.ResourceRequirements,
	nodeSelector map[string]string,
	env []v1.EnvVar,
	args []string,
	volumeMounts []v1.VolumeMount,
	volumes []v1.Volume) *v12.Deployment {

	allowPrivilegeEscalation := false
	readOnlyRootFilesystem := true
	var replicas int32 = 1

	tmpVolumes := []v1.Volume{
		v1.Volume{
			Name: "output",
			VolumeSource: v1.VolumeSource{
				EmptyDir: &v1.EmptyDirVolumeSource{},
			},
		},
		v1.Volume{
			Name: "tmp",
			VolumeSource: v1.VolumeSource{
				EmptyDir: &v1.EmptyDirVolumeSource{},
			},
		},
	}
	tmpVolumeMount := []v1.VolumeMount{
		v1.VolumeMount{
			Name:      "output",
			MountPath: "/opt/robot/output",
		},
		v1.VolumeMount{
			Name:      "tmp",
			MountPath: "/tmp",
		},
	}

	volumes = append(volumes, tmpVolumes...)
	volumeMounts = append(volumeMounts, tmpVolumeMount...)
	dc := &v12.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      utils.Robot,
			Namespace: namespace,
			Labels: map[string]string{
				utils.App:          utils.CassandraCluster,
				utils.Microservice: utils.Robot,
				utils.Name:         utils.Robot,
			},
		},
		Spec: v12.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					utils.Name: utils.Robot,
				},
			},
			Replicas: &replicas,
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: namespace,
					Labels: map[string]string{
						utils.Name: utils.Robot,
					},
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						v1.Container{
							Name:         utils.Robot,
							Image:        image,
							Env:          env,
							Resources:    resources,
							Args:         args,
							VolumeMounts: volumeMounts,
							SecurityContext: &v1.SecurityContext{
								Capabilities: &v1.Capabilities{
									Drop: []v1.Capability{"ALL"},
								},
								AllowPrivilegeEscalation: &allowPrivilegeEscalation,
								ReadOnlyRootFilesystem:   &readOnlyRootFilesystem,
							},
						},
					},
					Volumes:      volumes,
					NodeSelector: nodeSelector,
				},
			},
		},
	}

	return dc
}
