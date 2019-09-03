/*

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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// FooSpec defines the desired state of Foo
type FooSpec struct {
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Format:=string

	// the name of deployment which is owned by foo
	DeploymentName string `json:"deploymentName"`

	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Minimum=0

	// the replicas of deployment which is owned by foo
	Replicas *int32 `json:"replicas"`
}

// FooStatus defines the observed state of Foo
type FooStatus struct {
	// this is equal deployment.status.availableReplicas
	// +optional
	AvailableReplicas int32 `json:"availableReplicas"`
}

// +kubebuilder:object:root=true

// Foo is the Schema for the foos API
type Foo struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   FooSpec   `json:"spec,omitempty"`
	Status FooStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// FooList contains a list of Foo
type FooList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Foo `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Foo{}, &FooList{})
}
