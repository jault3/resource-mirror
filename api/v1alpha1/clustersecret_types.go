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

package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ClusterSecretSpec defines the desired state of ClusterSecret
type ClusterSecretSpec struct {
	// Data contains the secret data. Each key must consist of alphanumeric
	// characters, '-', '_' or '.'. Each value must be a base64 encoded string.
	// This is identical to the v1/Secret data field.
	// +optional
	Data map[string][]byte `json:"data,omitempty"`

	// The type of secret described in https://kubernetes.io/docs/concepts/configuration/secret/#secret-types.
	// Defaults to Opaque if not specified. This is identical to the v1/Secret
	// type field.
	// +optional
	Type corev1.SecretType `json:"type,omitempty"`
}

// ClusterSecretStatus defines the observed state of ClusterSecret
type ClusterSecretStatus struct {
	// Whether or not this ClusterSecret resource has been processed
	// and mirrored to the appropriate namespaces.
	Mirrored bool `json:"mirrored,omitempty"`

	// The timestamp when this ClusterSecret was last reconciled and
	// deployed to all namespaces.
	LastReconciled string `json:"lastReconciled,omitempty"`

	// A list of namespaces where this ClusterSecret was mirrored to.
	MirroredTo []string `json:"mirroredTo,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:scope=Cluster

// ClusterSecret is the Schema for the clustersecrets API
type ClusterSecret struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ClusterSecretSpec   `json:"spec,omitempty"`
	Status ClusterSecretStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ClusterSecretList contains a list of ClusterSecret
type ClusterSecretList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ClusterSecret `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ClusterSecret{}, &ClusterSecretList{})
}
