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

// SushrutAWSEC2Spec defines the desired state of SushrutAWSEC2
type SushrutAWSEC2Spec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of SushrutAWSEC2. Edit sushrutawsec2_types.go to remove/update
	Image string `json:"image,omitempty"`
}

// SushrutAWSEC2Status defines the observed state of SushrutAWSEC2
type SushrutAWSEC2Status struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// SushrutAWSEC2 is the Schema for the sushrutawsec2s API
type SushrutAWSEC2 struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SushrutAWSEC2Spec   `json:"spec,omitempty"`
	Status SushrutAWSEC2Status `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// SushrutAWSEC2List contains a list of SushrutAWSEC2
type SushrutAWSEC2List struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SushrutAWSEC2 `json:"items"`
}

func init() {
	SchemeBuilder.Register(&SushrutAWSEC2{}, &SushrutAWSEC2List{})
}
