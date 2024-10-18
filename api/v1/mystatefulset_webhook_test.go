/*
Copyright 2024.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	corev1 "k8s.io/api/core/v1"
)

var _ = Describe("MyStatefulSet Webhook", func() {

	Context("When creating MyStatefulSet under Defaulting Webhook", func() {
		It("Should fill in the default value if a required field is empty", func() {

			// TODO(user): Add your logic here

		})
	})

	Context("When creating MyStatefulSet under Validating Webhook", func() {
		It("Should deny if a required field is empty", func() {

			// TODO(user): Add your logic here

		})

		It("Should admit if all required fields are provided", func() {

			// TODO(user): Add your logic here

		})
	})

})

func TestMyStatefulSet_ValidateStatefulset(t *testing.T) {
	correctReplicas := new(int)
	wrongReplicas := new(int)
	*correctReplicas = 3
	*wrongReplicas = 0

	type fields struct {
		Spec MyStatefulSetSpec
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "OK",
			fields: fields{
				Spec: MyStatefulSetSpec{
					Replicas: correctReplicas,
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{Image: "redis"},
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Err replicas",
			fields: fields{
				Spec: MyStatefulSetSpec{
					Replicas: wrongReplicas,
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{Image: "redis"},
							},
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "Err image",
			fields: fields{
				Spec: MyStatefulSetSpec{
					Replicas: correctReplicas,
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{Image: ""},
							},
						},
					},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &MyStatefulSet{
				Spec: tt.fields.Spec,
			}
			if err := r.ValidateStatefulset(); (err != nil) != tt.wantErr {
				t.Errorf("MyStatefulSet.ValidateStatefulset() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
