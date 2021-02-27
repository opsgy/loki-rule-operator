/*
Copyright 2021.

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

package v1beta1

import (
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var lokirulelog = logf.Log.WithName("lokirule-resource")

func (r *LokiRule) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// +kubebuilder:webhook:path=/validate-logging-opsgy-com-v1beta1-lokirule,mutating=false,failurePolicy=fail,sideEffects=None,groups=logging.opsgy.com,resources=lokirules,verbs=create;update,versions=v1beta1,name=vlokirule.kb.io,admissionReviewVersions={v1,v1beta1}

var _ webhook.Validator = &LokiRule{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *LokiRule) ValidateCreate() error {
	_, err := r.ValidateExpressions()
	return err
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *LokiRule) ValidateUpdate(old runtime.Object) error {
	_, err := r.ValidateExpressions()
	return err
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *LokiRule) ValidateDelete() error {
	return nil
}
