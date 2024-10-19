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

package controller

import (
	"context"
	"fmt"
	devopsv1 "lucheng/api/v1"
	"lucheng/mocks/mock_client"
	"lucheng/mocks/mock_controller"
	"testing"

	"bou.ke/monkey"
	"github.com/go-logr/logr"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var _ = Describe("MyStatefulSet Controller", func() {
	Context("When reconciling a resource", func() {
		const resourceName = "test-resource"

		ctx := context.Background()

		typeNamespacedName := types.NamespacedName{
			Name:      resourceName,
			Namespace: "default", // TODO(user):Modify as needed
		}
		mystatefulset := &devopsv1.MyStatefulSet{}

		BeforeEach(func() {
			By("creating the custom resource for the Kind MyStatefulSet")
			err := k8sClient.Get(ctx, typeNamespacedName, mystatefulset)
			if err != nil && errors.IsNotFound(err) {
				resource := &devopsv1.MyStatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      resourceName,
						Namespace: "default",
					},
					// TODO(user): Specify other spec details if needed.
				}
				Expect(k8sClient.Create(ctx, resource)).To(Succeed())
			}
		})

		AfterEach(func() {
			// TODO(user): Cleanup logic after each test, like removing the resource instance.
			resource := &devopsv1.MyStatefulSet{}
			err := k8sClient.Get(ctx, typeNamespacedName, resource)
			Expect(err).NotTo(HaveOccurred())

			By("Cleanup the specific resource instance MyStatefulSet")
			Expect(k8sClient.Delete(ctx, resource)).To(Succeed())
		})
		It("should successfully reconcile the resource", func() {
			By("Reconciling the created resource")
			controllerReconciler := &MyStatefulSetReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())
			// TODO(user): Add more specific assertions depending on your controller's reconciliation logic.
			// Example: If you expect a certain status condition after reconciliation, verify it here.
		})
	})
})

type PatchObj struct {
	PatchFunc  interface{}
	TargetFunc interface{}
}

func TestMyStatefulSetReconciler_Reconcile(t *testing.T) {
	// Setup
	mock_ctrl := gomock.NewController(t)
	defer mock_ctrl.Finish()
	mock_mgr := mock_controller.NewMockMgr(mock_ctrl)
	mock_client := mock_client.NewMockClient(mock_ctrl)
	mock_client.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(errors.NewBadRequest("err"))
	mock_client.EXPECT().Update(gomock.Any(), gomock.Any()).AnyTimes()

	monkey.Patch(NewMgr, func(*MyStatefulSetReconciler, *devopsv1.MyStatefulSet, logr.Logger) Mgr {
		return mock_mgr
	})

	// Init args
	ctx := context.Background()
	req := ctrl.Request{}

	tests := []struct {
		name      string
		wantErr   bool
		patchList []*PatchObj
	}{
		{
			name:    "Err fetch failed",
			wantErr: true,
			patchList: []*PatchObj{
				{
					PatchFunc: client.IgnoreNotFound,
					TargetFunc: func(error) error {
						return errors.NewBadRequest("err")
					},
				},
			},
		},
	}
	for _, tt := range tests {
		// Monkey patch
		for _, obj := range tt.patchList {
			monkey.Patch(obj.PatchFunc, obj.TargetFunc)
		}

		t.Run(tt.name, func(t *testing.T) {
			r := &MyStatefulSetReconciler{
				Client: mock_client,
				Scheme: runtime.NewScheme(),
			}
			_, err := r.Reconcile(ctx, req)
			if (err != nil) != tt.wantErr {
				fmt.Println(err)
				t.Errorf("MyStatefulSetReconciler.Reconcile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestMgrTeardown(t *testing.T) {
	// Setup mock client
	mock_ctrl := gomock.NewController(t)
	defer mock_ctrl.Finish()
	mock_client := mock_client.NewMockClient(mock_ctrl)

	mock_client.EXPECT().List(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(nil)
	mock_client.EXPECT().Scheme().AnyTimes()

	// Mgr and ctx
	replicas := new(int)
	*replicas = 3
	ctx := context.Background()
	rec := MyStatefulSetReconciler{
		Client: mock_client,
		Scheme: mock_client.Scheme(),
	}
	set := devopsv1.MyStatefulSet{
		Spec: devopsv1.MyStatefulSetSpec{
			Replicas: replicas,
		},
	}
	logger := log.FromContext(ctx)
	mgr := NewMgr(&rec, &set, logger)

	// Testcase
	tests := []struct {
		name      string
		wantErr   bool
		patchList []*PatchObj
	}{
		{
			name:    "empty resource",
			wantErr: false,
			patchList: []*PatchObj{
				{
					PatchFunc: mgr.(*SetMgr).DeletePod,
					TargetFunc: func(context.Context, int) error {
						return errors.NewBadRequest("delete pod error")
					},
				},
				{
					PatchFunc: mgr.(*SetMgr).ReleasePvc,
					TargetFunc: func(context.Context, int) error {
						return errors.NewBadRequest("release pvc error")
					},
				},
			},
		},
	}

	for _, tt := range tests {
		// Monkey patch
		for _, obj := range tt.patchList {
			monkey.Patch(obj.PatchFunc, obj.TargetFunc)
		}

		t.Run(tt.name, func(t *testing.T) {
			err := mgr.Teardown(ctx)
			if (err != nil) != tt.wantErr {
				fmt.Println(err)
				t.Errorf("MyStatefulSetReconciler.Reconcile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
