package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// EnvDeploymentSpec defines the desired state of EnvDeployment
type EnvDeploymentSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file

	// EnvironmentRef points to the deployment config
	EnvironmentRef EnvironmentRef `json:"environmentRef" protobuf:"bytes,1,opt,name=environmentRef"`

	// DeploymentRef points to the deployment config
	DeploymentRef DeploymentRef `json:"deploymentRef" protobuf:"bytes,2,opt,name=deploymentRef"`
}

// EnvironmentRef contains information that points to the environment
type EnvironmentRef struct {
	// APIGroup is the group for the resource being referenced
	APIGroup string `json:"apiGroup" protobuf:"bytes,1,opt,name=apiGroup"`

	// Kind is the type of resource being referenced
	Kind string `json:"kind" protobuf:"bytes,2,opt,name=kind"`

	// Name is the name of resource being referenced
	Name string `json:"name" protobuf:"bytes,3,opt,name=name"`
}

// DeploymentRef contains information that points to the deployment config
type DeploymentRef struct {
	// APIGroup is the group for the resource being referenced
	APIGroup string `json:"apiGroup" protobuf:"bytes,1,opt,name=apiGroup"`

	// Kind is the type of resource being referenced
	Kind string `json:"kind" protobuf:"bytes,2,opt,name=kind"`

	// Name is the name of resource being referenced
	Name string `json:"name" protobuf:"bytes,3,opt,name=name"`
}

// EnvDeploymentStatus defines the observed state of EnvDeployment
type EnvDeploymentStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// EnvDeployment is the Schema for the envdeployments API
// +k8s:openapi-gen=true
type EnvDeployment struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   EnvDeploymentSpec   `json:"spec,omitempty"`
	Status EnvDeploymentStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// EnvDeploymentList contains a list of EnvDeployment
type EnvDeploymentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []EnvDeployment `json:"items"`
}

func init() {
	SchemeBuilder.Register(&EnvDeployment{}, &EnvDeploymentList{})
}
