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

package controllers

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"gopkg.in/yaml.v2"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	loggingv1beta1 "github.com/opsgy/loki-rule-operator/api/v1beta1"
)

// GlobalLokiRuleReconciler reconciles a GlobalLokiRule object
type GlobalLokiRuleReconciler struct {
	client.Client
	Log                     logr.Logger
	Scheme                  *runtime.Scheme
	Clientset               *kubernetes.Clientset
	RulesConfigMapName      string
	RulesConfigMapNamespace string
}

// +kubebuilder:rbac:groups=logging.opsgy.com,resources=globallokirules,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=logging.opsgy.com,resources=globallokirules/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=logging.opsgy.com,resources=globallokirules/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the GlobalLokiRule object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.7.0/pkg/reconcile
func (r *GlobalLokiRuleReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = r.Log.WithValues("globallokirule", req.NamespacedName)

	// your logic here
	fileName := req.Namespace + "-" + req.Name + ".yml"
	lokiRule := &loggingv1beta1.GlobalLokiRule{}
	err := r.Get(ctx, req.NamespacedName, lokiRule)
	if err != nil {
		if errors.IsNotFound(err) {
			// Remove item from configmap
			cm, err := r.Clientset.CoreV1().ConfigMaps(r.RulesConfigMapNamespace).Get(ctx, r.RulesConfigMapName, metav1.GetOptions{})
			if err != nil {
				if errors.IsNotFound(err) {
					// do nothing
					return ctrl.Result{}, nil
				}
				return ctrl.Result{Requeue: true}, err
			}

			if _, ok := cm.Labels["app.kubernetes.io/managed-by"]; !ok {
				return ctrl.Result{Requeue: true}, fmt.Errorf("ConfigMap %s/%s is missing label app.kubernetes.io/managed-by", r.RulesConfigMapNamespace, r.RulesConfigMapName)
			} else if cm.Labels["app.kubernetes.io/managed-by"] != "loki-rule-operator" {
				return ctrl.Result{Requeue: true}, fmt.Errorf("ConfigMap %s/%s is managed by someone else", r.RulesConfigMapNamespace, r.RulesConfigMapName)
			}

			if cm.Data != nil {
				_, exists := cm.Data[fileName]
				if exists {
					delete(cm.Data, fileName)
					_, err = r.Clientset.CoreV1().ConfigMaps(r.RulesConfigMapNamespace).Update(ctx, cm, metav1.UpdateOptions{})
					if err != nil {
						return ctrl.Result{Requeue: true}, err
					}
				}
			}
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	// Evaluate rules
	spec, err := lokiRule.ValidateExpressions()
	if err != nil {
		lokiRule.Status.Valid = false
		lokiRule.Status.Message = err.Error()
		r.Client.Status().Update(context.TODO(), lokiRule)
		return ctrl.Result{}, nil
	}
	if !lokiRule.Status.Valid {
		lokiRule.Status.Valid = true
		lokiRule.Status.Message = ""
		r.Client.Status().Update(context.TODO(), lokiRule)
	}
	lokiRule.Spec = *spec

	data, err := yaml.Marshal(&lokiRule.Spec)
	if err != nil {
		lokiRule.Status.Message = err.Error()
		r.Client.Status().Update(context.TODO(), lokiRule)
		return ctrl.Result{}, err
	}

	// Update ConfigMap
	cm, err := r.Clientset.CoreV1().ConfigMaps(r.RulesConfigMapNamespace).Get(ctx, r.RulesConfigMapName, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			labelMap := make(map[string]string)
			labelMap["app.kubernetes.io/managed-by"] = "loki-rule-operator"
			dataMap := make(map[string]string)
			dataMap[fileName] = string(data)

			cm := &v1.ConfigMap{
				TypeMeta: metav1.TypeMeta{
					Kind:       "ConfigMap",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      r.RulesConfigMapName,
					Namespace: r.RulesConfigMapNamespace,
					Labels:    labelMap,
				},
				Data: dataMap,
			}
			_, err := r.Clientset.CoreV1().ConfigMaps(r.RulesConfigMapNamespace).Create(ctx, cm, metav1.CreateOptions{})
			if err != nil {
				return ctrl.Result{Requeue: true}, err
			}
		} else {
			return ctrl.Result{Requeue: true}, err
		}
	} else {
		if _, ok := cm.Labels["app.kubernetes.io/managed-by"]; !ok {
			return ctrl.Result{Requeue: true}, fmt.Errorf("ConfigMap %s/%s is missing label app.kubernetes.io/managed-by", r.RulesConfigMapNamespace, r.RulesConfigMapName)
		} else if cm.Labels["app.kubernetes.io/managed-by"] != "loki-rule-operator" {
			return ctrl.Result{Requeue: true}, fmt.Errorf("ConfigMap %s/%s is managed by someone else", r.RulesConfigMapNamespace, r.RulesConfigMapName)
		}

		if cm.Data == nil {
			cm.Data = make(map[string]string)
		}
		cm.Data[fileName] = string(data)
		_, err := r.Clientset.CoreV1().ConfigMaps(r.RulesConfigMapNamespace).Update(ctx, cm, metav1.UpdateOptions{})
		if err != nil {
			return ctrl.Result{Requeue: true}, err
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *GlobalLokiRuleReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&loggingv1beta1.GlobalLokiRule{}).
		Complete(r)
}
