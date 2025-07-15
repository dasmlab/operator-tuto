package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// SimpleCRDSpec defines the desired state of SimpleCRD
type SimpleCRDSpec struct {
	NumPods     int    `json:"numPods,omitempty"`
	LastChecked string `json:"lastChecked,omitempty"`
	NumIter     int    `json:"numIter,omitempty"`
}

// SimpleCRDStatus defines the observed state of SimpleCRD.
type SimpleCRDStatus struct {
	Waiting   bool `json:"waitingTillNext,omitempty"`
	Completed bool `json:"completedTask,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// SimpleCRD is the Schema for the simplecrds API
type SimpleCRD struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is a standard object metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty,omitzero"`

	// spec defines the desired state of SimpleCRD
	// +required
	Spec SimpleCRDSpec `json:"spec"`

	// status defines the observed state of SimpleCRD
	// +optional
	Status SimpleCRDStatus `json:"status,omitempty,omitzero"`
}

// +kubebuilder:object:root=true

// SimpleCRDList contains a list of SimpleCRD
type SimpleCRDList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SimpleCRD `json:"items"`
}

func init() {
	SchemeBuilder.Register(&SimpleCRD{}, &SimpleCRDList{})
}
