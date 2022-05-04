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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// LokiRuleSpec defines the desired state of LokiRule
type LokiRuleSpec struct {
	Groups []*LokiRuleGroup `json:"groups,omitempty" yaml:"groups"`
}

type LokiRuleGroup struct {
	Interval string           `json:"interval,omitempty" yaml:"interval,omitempty"`
	Name     string           `json:"name,omitempty" yaml:"name"`
	Rules    []*LokiGroupRule `json:"rules,omitempty" yaml:"rules"`
}

type LokiGroupRule struct {
	Alert       string            `json:"alert,omitempty" yaml:"alert,omitempty"`
	Record      string            `json:"record,omitempty" yaml:"record,omitempty"`
	Expr        string            `json:"expr,omitempty" yaml:"expr"`
	For         string            `json:"for,omitempty" yaml:"for,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty" yaml:"annotations,omitempty"`
	Labels      map[string]string `json:"labels,omitempty" yaml:"labels,omitempty"`
}

// LokiRuleStatus defines the observed state of LokiRule
type LokiRuleStatus struct {
	Valid   bool   `json:"valid"`
	Message string `json:"message,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// LokiRule is the Schema for the lokirules API
type LokiRule struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   LokiRuleSpec   `json:"spec,omitempty"`
	Status LokiRuleStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// LokiRuleList contains a list of LokiRule
type LokiRuleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []LokiRule `json:"items"`
}

func init() {
	SchemeBuilder.Register(&LokiRule{}, &LokiRuleList{})
}
