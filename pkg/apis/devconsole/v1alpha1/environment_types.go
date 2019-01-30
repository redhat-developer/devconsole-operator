package v1alpha1

import (
	clusterv1alpha1 "github.com/alexeykazakov/devconsole/pkg/apis/clusterregistry/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// EnvironmentSpec defines the desired state of Environment
type EnvironmentSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of Environment
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file

	// Type contains the type of the environment.
	// should be treated as "deploy" if nil
	// +optional
	Type string `json:"name,omitempty" protobuf:"bytes,2,opt,name=type"`

	// Cluster contains information about a cluster.
	// If empty then the current cluster should be used.
	// +optional
	Cluster clusterv1alpha1.Cluster `json:"cluster,omitempty" protobuf:"bytes,2,opt,name=cluster"`

	// Namespace contains the namespace of the environment.
	// If empty then the current namespace should be used.
	// +optional
	Namespace string `json:"namespace,omitempty" protobuf:"bytes,3,opt,name=namespace"`
}

// EnvironmentStatus defines the observed state of Environment
type EnvironmentStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Environment is the Schema for the environments API
// +k8s:openapi-gen=true
type Environment struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   EnvironmentSpec   `json:"spec,omitempty"`
	Status EnvironmentStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// EnvironmentList contains a list of Environment
type EnvironmentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Environment `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Environment{}, &EnvironmentList{})
}
