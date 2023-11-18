package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ExecuterSpec defines the desired state of Executer
type ExecuterSpec struct {
	// Name is the name of the Executer
	// +kubebuilder:validation:Required
	Name string `json:"name,omitempty"`

	// Image is the name of the image to be used for executer
	// +kubebuilder:validation:Required
	Image string `json:"image,omitempty"`

	// Commands is the command to be run inside the container
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinItems:=1
	Commands []string `json:"commands,omitempty"`

	// Replication is the replicas for the executer
	// +kubebuilder:validation:Optional
	Replication int32 `json:"replication,omitempty"`
}

type Phase string

const (
	PhaseUnknown  Phase = "Unknown"
	PhaseIdle     Phase = "Idle"
	PhaseCreating Phase = "Creating"
	PhaseCreated  Phase = "Created"
	PhaseUpdating Phase = "Updating"
	PhaseFailed   Phase = "Failed"
)

// ExecuterStatus defines the observed state of Executer
type ExecuterStatus struct {
	Phase Phase `json:"phase,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Executer is the Schema for the executers API
type Executer struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ExecuterSpec   `json:"spec,omitempty"`
	Status ExecuterStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ExecuterList contains a list of Executer
type ExecuterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Executer `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Executer{}, &ExecuterList{})
}
