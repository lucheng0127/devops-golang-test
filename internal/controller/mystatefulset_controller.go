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

	"github.com/go-logr/logr"
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
	Sync(context.Context) error
	Teardown(context.Context) error
}

type SetMgr struct {
	Set    *devopsv1.MyStatefulSet
	Rec    *MyStatefulSetReconciler
	Logger logr.Logger
}

func NewMgr(r *MyStatefulSetReconciler, set *devopsv1.MyStatefulSet, logger logr.Logger) Mgr {
	return &SetMgr{Set: set, Rec: r, Logger: logger}
}

func (mgr *SetMgr) CreatePod(ctx context.Context, idx, timeout int) (*corev1.Pod, error) {
	// TODO(shawn): Implement it
	return nil, nil
}

func (mgr *SetMgr) ClaimPvc(ctx context.Context, idx int) (*corev1.PersistentVolumeClaim, error) {
	// TODO(shawn): Implement it
	return nil, nil
}

func (mgr *SetMgr) DeletePod(ctx context.Context, idx int) error {
	// TODO(shawn): Implement it
	return nil
}

func (mgr *SetMgr) ReleasePvc(ctx context.Context, idx int) error {
	// TODO(shawn): Implement it
	return nil
}

func (mgr *SetMgr) podNameByIdx(idx int) string {
	return fmt.Sprintf("pod-%s-%d", mgr.Set.Name, idx)
}

func (mgr *SetMgr) pvcNameByIdx(idx int) string {
	return fmt.Sprintf("pvc-%s-%d", mgr.Set.Name, idx)
}

func (mgr *SetMgr) Sync(ctx context.Context) error {
	// List ready pod
	setPods := &corev1.PodList{}
	query := client.MatchingLabels{"mangledBy": mgr.Set.Name}
	if err := mgr.Rec.Client.List(ctx, setPods, client.InNamespace(mgr.Set.Namespace), query); client.IgnoreAlreadyExists(err) != nil {
		return err
	}

	setPodNum := 0
	for _, pod := range setPods.Items {
		if pod.Status.Phase == corev1.PodRunning || pod.Status.Phase == corev1.PodPending {
			setPodNum++
		}
	}

	// Shrink or add pods according to replicas
	if setPodNum == *mgr.Set.Spec.Replicas {
		// Do nothing and return
		mgr.Set.Status.PodIdx = setPodNum - 1
		return nil
	} else if setPodNum > *mgr.Set.Spec.Replicas {
		// Shrink, delete pod but keep pvc
		for idx := *mgr.Set.Spec.Replicas; idx > setPodNum; idx-- {
			mgr.Logger.Info(fmt.Sprintf("try to delete MyStatefulset %s pod with idx %d", mgr.Set.Name, idx))

			if err := mgr.DeletePod(ctx, idx); err != nil {
				return err
			}

			mgr.Logger.Info(fmt.Sprintf("delete MyStatefulset %s pod with idx %d succeed", mgr.Set.Name, idx))
		}

		return nil
	} else {
		// Add, create pod and pvc
		for idx := setPodNum; idx < *mgr.Set.Spec.Replicas; idx++ {
			// Claim pvc
			_, err := mgr.ClaimPvc(ctx, idx)
			if client.IgnoreAlreadyExists(err) != nil {
				return err
			}
			// mgr.Logger.Info(fmt.Sprintf("claim pvc %s naemspace %s succeed", pvc.Name, pvc.Namespace))
			mgr.Logger.Info(fmt.Sprintf("claim pvc %d succeed", idx))

			// Create pod
			_, err = mgr.CreatePod(ctx, idx, *mgr.Set.Spec.GracePeriod)
			if err != nil {
				return err
			}

			// mgr.Logger.Info(fmt.Sprintf("create pod %s naemspace %s succeed", pod.Name, pod.Namespace))
			mgr.Logger.Info(fmt.Sprintf("create pod %d succeed", idx))
		}

		return nil
	}
}

func (mgr *SetMgr) Teardown(ctx context.Context) error {
	// List all pods
	setPods := &corev1.PodList{}
	query := client.MatchingLabels{"mangledBy": mgr.Set.Name}
	if err := mgr.Rec.Client.List(ctx, setPods, client.InNamespace(mgr.Set.Namespace), query); client.IgnoreAlreadyExists(err) != nil {
		return err
	}

	errReturn := false

	// Delete pods reverse
	for idx := len(setPods.Items); idx >= 0; idx-- {
		if err := mgr.DeletePod(ctx, idx); client.IgnoreNotFound(err) != nil {
			errReturn = true
			mgr.Logger.Error(err, fmt.Sprintf("delete MyStatefulset %s pod %d", mgr.Set.Name, idx))
		}
	}

	// List all pvc, beacuse pvc maybe more than pods when update statefulset replicas
	setPvcs := &corev1.PersistentVolumeClaimList{}
	if err := mgr.Rec.Client.List(ctx, setPvcs, client.InNamespace(mgr.Set.Namespace), query); client.IgnoreAlreadyExists(err) != nil {
		return err
	}

	// Release pvc
	for idx := len(setPvcs.Items); idx >= 0; idx-- {
		if err := mgr.ReleasePvc(ctx, idx); client.IgnoreNotFound(err) != nil {
			errReturn = true
			mgr.Logger.Error(err, fmt.Sprintf("release MyStatefulset %s pvc %d", mgr.Set.Name, idx))
		}
	}

	if errReturn {
		return fmt.Errorf("teardown MyStatefulSet %s error", mgr.Set.Name)
	}

	return nil
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

	// Init mgr
	mgr := NewMgr(r, &mySet, logger)

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

			// Call mgr teardown pods and pvcs for statefulset
			if err := mgr.Teardown(ctx); err != nil {
				logger.Error(err, fmt.Sprintf("delete MyStatefulSet %s namespace %s", mySet.ObjectMeta.Name, mySet.ObjectMeta.Namespace))
				return ctrl.Result{}, err
			}

			logger.Info(fmt.Sprintf("delete MyStatefulSet %s in namespace %s finished", mySet.ObjectMeta.Name, mySet.ObjectMeta.Namespace))
		}

		controllerutil.RemoveFinalizer(&mySet, FZNAME)
		logger.Info("Delete MyStatefulSet finished")

		needUpdate = true
		return ctrl.Result{}, nil
	}

	// Call mgr setup pods and pvcs for statefulset utils all ready or timeout
	if err := mgr.Sync(ctx); err != nil {
		logger.Error(err, fmt.Sprintf("sync MyStatefulSet %s namespace %s resource", mySet.ObjectMeta.Name, mySet.ObjectMeta.Namespace))

		mySet.Status.ErrReason = err.Error()

		needUpdate = true
		return ctrl.Result{}, err
	}

	logger.Info(fmt.Sprintf("sync MyStatefulSet %s in namespace %s resource finished", mySet.ObjectMeta.Name, mySet.ObjectMeta.Namespace))
	needUpdate = true
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *MyStatefulSetReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&devopsv1.MyStatefulSet{}).
		Complete(r)
}
