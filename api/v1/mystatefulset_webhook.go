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
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// log is for logging in this package.
var mystatefulsetlog = logf.Log.WithName("mystatefulset-resource")

// SetupWebhookWithManager will setup the manager to manage the webhooks
func (r *MyStatefulSet) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// +kubebuilder:webhook:path=/mutate-devops-github-com-v1-mystatefulset,mutating=true,failurePolicy=fail,sideEffects=None,groups=devops.github.com,resources=mystatefulsets,verbs=create;update;delete,versions=v1,name=mmystatefulset.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &MyStatefulSet{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *MyStatefulSet) Default() {
	mystatefulsetlog.Info("default", "name", r.Name)

	// Check grace period, if not set it means do not check pod ready set it to -1,
	// if set when create pods for statefulset, if timeout for pod ready, set statefulset status to error
	if r.Spec.GracePeriod == nil {
		r.Spec.GracePeriod = new(int)
		*r.Spec.GracePeriod = -1
	}
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
// NOTE: The 'path' attribute must follow a specific pattern and should not be modified directly here.
// Modifying the path for an invalid path can cause API server errors; failing to locate the webhook.
// +kubebuilder:webhook:path=/validate-devops-github-com-v1-mystatefulset,mutating=false,failurePolicy=fail,sideEffects=None,groups=devops.github.com,resources=mystatefulsets,verbs=create;update;delete,versions=v1,name=vmystatefulset.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &MyStatefulSet{}

func (r *MyStatefulSet) ValidateReplicas() *field.Error {
	if *r.Spec.Replicas < 1 {
		return field.Invalid(
			field.NewPath("spec").Child("replicas"),
			*r.Spec.Replicas,
			"replicas must large than 1",
		)
	}

	return nil
}

func (r *MyStatefulSet) ValidatePodImage() *field.Error {
	for _, c := range r.Spec.Template.Spec.Containers {
		if c.Image == "" {
			return field.Required(
				field.NewPath("sepc").Child("template").Child("spec").Child("containers").Child("image"),
				"image can't be empty",
			)
		}
	}

	return nil
}

func (r *MyStatefulSet) ValidateStatefulset() error {
	var errs field.ErrorList

	if err := r.ValidateReplicas(); err != nil {
		errs = append(errs, err)
	}

	if err := r.ValidatePodImage(); err != nil {
		errs = append(errs, err)
	}

	if len(errs) == 0 {
		return nil
	}

	return apierrors.NewInvalid(
		schema.GroupKind{Group: "devops", Kind: "MyStatefulSet"},
		r.Name,
		errs,
	)
}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *MyStatefulSet) ValidateCreate() (admission.Warnings, error) {
	mystatefulsetlog.Info("validate create", "name", r.Name)

	return nil, r.ValidateStatefulset()
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *MyStatefulSet) ValidateUpdate(old runtime.Object) (admission.Warnings, error) {
	mystatefulsetlog.Info("validate update", "name", r.Name)

	return nil, r.ValidateStatefulset()
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *MyStatefulSet) ValidateDelete() (admission.Warnings, error) {
	mystatefulsetlog.Info("validate delete", "name", r.Name)

	// TODO(shawn): Implement it if needed
	return nil, nil
}
