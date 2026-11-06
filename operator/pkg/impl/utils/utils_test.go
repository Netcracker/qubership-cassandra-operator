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

package utils

import (
	"reflect"
	"testing"

	"github.com/Netcracker/qubership-cassandra-operator/api/v1alpha1"
)

func TestFilterDC(t *testing.T) {
	type args struct {
		input  []*v1alpha1.DataCenter
		filter func(dc *v1alpha1.DataCenter) bool
	}
	dcsAll := []*v1alpha1.DataCenter{
		{
			Deploy:   true,
			Name:     "dc1",
			Replicas: 3,
		},
		{
			Deploy:   true,
			Name:     "dc2",
			Replicas: 3,
		},
	}

	dcsOne := []*v1alpha1.DataCenter{
		{
			Deploy:   true,
			Name:     "dc1",
			Replicas: 3,
		},
		{
			Deploy:   false,
			Name:     "dc2",
			Replicas: 3,
		},
	}

	tests := []struct {
		name string
		args args
		want []*v1alpha1.DataCenter
	}{
		{
			name: "Test All DC deployed",
			args: args{
				input:  dcsAll,
				filter: func(dc *v1alpha1.DataCenter) bool { return dc.Deploy },
			},
			want: dcsAll,
		},
		{
			name: "Test One DC deployed",
			args: args{
				input:  dcsOne,
				filter: func(dc *v1alpha1.DataCenter) bool { return dc.Deploy },
			},
			want: []*v1alpha1.DataCenter{
				{
					Deploy:   true,
					Name:     "dc1",
					Replicas: 3,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// mass := NewTypedStream(dcsAll, &v13.DataCenter{}).Slice()
			// nn := make([]*v13.DataCenter, len(mass))
			// for i, mas := range mass {
			// 	nn[i] = mas.(*v13.DataCenter)
			// }
			if got := FilterDC(tt.args.input, tt.args.filter); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FilterDC() = %v, want %v", got, tt.want)
			}
		})
	}
}
