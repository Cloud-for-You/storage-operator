/*
Copyright 2023.

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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

const (
	PhasePending        = "Pending"
	PhaseBound          = "Bound"
	PhaseUnknown        = "Unknown"
	PhaseError          = "Error"
	AutomationCompleted = "Completed"
	AutomationError     = "Error"
	AutomationRunning   = "Running"
)

// NfsSpec defines the desired state of Nfs
type NfsSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// +kubebuilder:validation:Required
	// +kubebuilder:subresource:status
	Server string `json:"server"`
	// +kubebuilder:validation:Required
	// +kubebuilder:subresource:status
	Path string `json:"path"`
	// Capacity must follow the Kubernetes resource quantity format
	// Example: 10Gi, 500Mi, etc.
	// +kubebuilder:validation:Pattern=`^([0-9]+)(Ei|Pi|Ti|Gi|Mi|Ki|e|p|t|g|m|k)?$`
	// +kubebuilder:default:="1Gi"
	Capacity string `json:"capacity,omitempty"`
}

// NfsStatus defines the observed state of Nfs
type NfsStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Phase      string `json:"phase,omitempty"`
	PVCName    string `json:"pvcName,omitempty"`
	Automation string `json:"automation,omitempty"`
	Message    string `json:"message,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="STATUS",type=string,JSONPath=`.status.phase`
//+kubebuilder:printcolumn:name="CLAIM",type=string,JSONPath=`.status.pvcName`
//+kubebuilder:printcolumn:name="CAPACITY",type=string,JSONPath=`.spec.capacity`
//+kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
//+kubebuilder:printcolumn:name="MESSAGE",type=string,JSONPath=`.status.message`

// Nfs is the Schema for the nfs API
type Nfs struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NfsSpec   `json:"spec,omitempty"`
	Status NfsStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// NfsList contains a list of Nfs
type NfsList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Nfs `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Nfs{}, &NfsList{})
}
