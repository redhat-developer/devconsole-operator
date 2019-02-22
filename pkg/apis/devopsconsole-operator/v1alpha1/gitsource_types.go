package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// GitSourceSpec defines the desired state of GitSource
// +k8s:openapi-gen=true
type GitSourceSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book.kubebuilder.io/beyond_basics/generating_crd.html

	// URL of the git repo
	URL string `json:"url"`

	// Ref is a git reference. Optional. "master" is used by default.
	Ref string `json:"ref,omitempty"`

	// ContextDir is a path to subfolder in the repo. Optional.
	ContextDir string `json:"contextDir,omitempty"`

	// HttpProxy is optional.
	HttpProxy string `json:"httpProxy,omitempty"`

	// HttpsProxy is optional.
	HttpsProxy string `json:"httpsProxy,omitempty"`

	// NoProxy can be used to specify domains for which no proxying should be performed. Optional.
	NoProxy string `json:"noProxy,omitempty"`

	// SourceSecret is the name of the secret that contains credentials to access the git repo. Optional.
	SourceSecret string `json:"sourceSecret,omitempty"`
}

// GitSourceStatus defines the observed state of GitSource
// +k8s:openapi-gen=true
type GitSourceStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book.kubebuilder.io/beyond_basics/generating_crd.html
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// GitSource is the Schema for the gitsources API
// +k8s:openapi-gen=true
type GitSource struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GitSourceSpec   `json:"spec,omitempty"`
	Status GitSourceStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// GitSourceList contains a list of GitSource
type GitSourceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GitSource `json:"items"`
}

func init() {
	SchemeBuilder.Register(&GitSource{}, &GitSourceList{})
}
