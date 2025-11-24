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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// IMPORTANT: Run "make manifests" to regenerate code after modifying this file
// NOTE: json tags are required. Any new fields must have json tags.

// GuestBookSpec defines the desired state of GuestBook
type GuestBookSpec struct {
	// Replicas is the number of guestbook instances
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=10
	// +kubebuilder:default=1
	Replicas int32 `json:"replicas,omitempty"`

	// WelcomeMessage is displayed on the guestbook page
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:default="Welcome to our Guestbook!"
	WelcomeMessage string `json:"welcomeMessage,omitempty"`
}

// GuestBookStatus defines the observed state of GuestBook
type GuestBookStatus struct {
	// AvailableReplicas is the number of running replicas
	AvailableReplicas int32 `json:"availableReplicas"`

	// URL is the service endpoint
	URL string `json:"url,omitempty"`

	// Conditions represent the latest observations of the GuestBook state
	// +patchMergeKey=type
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=gb
// +kubebuilder:printcolumn:name="Replicas",type=integer,JSONPath=`.spec.replicas`
// +kubebuilder:printcolumn:name="Available",type=integer,JSONPath=`.status.availableReplicas`
// +kubebuilder:printcolumn:name="Message",type=string,JSONPath=`.spec.welcomeMessage`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// GuestBook is the Schema for the guestbooks API
type GuestBook struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GuestBookSpec   `json:"spec,omitempty"`
	Status GuestBookStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// GuestBookList contains a list of GuestBook
type GuestBookList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GuestBook `json:"items"`
}

func init() {
	SchemeBuilder.Register(&GuestBook{}, &GuestBookList{})
}
