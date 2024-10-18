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
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	devopsv1 "lucheng/api/v1"
)

const (
	FZNAME = "github.com/finlizer"
)

// MyStatefulSetReconciler reconciles a MyStatefulSet object
type MyStatefulSetReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

func ignoreErrs(err error) error {
	if err == nil {
		return nil
	}

	if strings.Contains(err.Error(), "resource name may not be empty") {
		return nil
	}

	return err
}

// StatefulsetMgr
// A interface, than can create pvc and pod for statefulset
// Use interface so can be mock by gomock

type Mgr interface {
	CreatePod(context.Context, int, int) (*corev1.Pod, error)
	ClaimPvc(context.Context, int) (*corev1.PersistentVolumeClaim, error)
	DeletePod(context.Context, int) error
	ReleasePvc(context.Context, int) error
}

// +kubebuilder:rbac:groups=devops.github.com,resources=mystatefulsets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=devops.github.com,resources=mystatefulsets/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=devops.github.com,resources=mystatefulsets/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the MyStatefulSet object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.19.0/pkg/reconcile
func (r *MyStatefulSetReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// Get current statefulset
	var mySet devopsv1.MyStatefulSet

	if err := r.Get(ctx, req.NamespacedName, &mySet); client.IgnoreNotFound(err) != nil {
		logger.Error(err, "unable to fetch MyStatefulset")
		return ctrl.Result{}, err
	}

	// Update statefulset before return
	needUpdate := false
	defer func() {
		if needUpdate {
			if err := r.Update(ctx, &mySet); ignoreErrs(err) != nil {
				logger.Error(err, "update MyStatefulSet")
			}
		}
	}()

	// TODO(shawn): Init pod manager and pvc manager

	// Handle delete statefulset
	if mySet.ObjectMeta.DeletionTimestamp.IsZero() {
		// Add finalizer for statefulset
		if !controllerutil.ContainsFinalizer(&mySet, FZNAME) {
			controllerutil.AddFinalizer(&mySet, FZNAME)

			needUpdate = true
			return ctrl.Result{}, nil
		}
	} else {
		// Handle delete statefulset
		if controllerutil.ContainsFinalizer(&mySet, FZNAME) {
			// Delete pods reversed order
			logger.Info(fmt.Sprintf("try to delete MyStatefulSet %s in namespace %s", mySet.ObjectMeta.Name, mySet.ObjectMeta.Namespace))

			// TODO(shawn): Call mgr teardown pods and pvcs for statefulset
		}

		controllerutil.RemoveFinalizer(&mySet, FZNAME)
		logger.Info("Delete MyStatefulSet finished")

		needUpdate = true
		return ctrl.Result{}, nil
	}

	// Check status of statefulset pods, if all ready do nothing and return

	// TODO(shawn): Call mgr setup pods and pvcs for statefulset utils all ready or timeout

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *MyStatefulSetReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&devopsv1.MyStatefulSet{}).
		Complete(r)
}
