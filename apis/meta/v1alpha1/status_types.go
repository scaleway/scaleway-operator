package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Status defines the observed state of a Scaleway resource
type Status struct {
	Conditions []Condition `json:"conditions,omitempty"`
}

// Condition contains details for the current condition of this Scaleway resource.
type Condition struct {
	// Type is the type of the condition.
	Type ConditionType `json:"type,omitempty"`
	// Status is the status of the condition.
	// Can be True, False, Unknown.
	Status corev1.ConditionStatus `json:"status,omitempty"`
	// Last time we probed the condition.
	LastProbeTime metav1.Time `json:"lastProbeTime,omitempty"`
	// Last time the condition transitioned from one status to another.
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty"`
	// Unique, one-word, CamelCase reason for the condition's last transition.
	Reason string `json:"reason,omitempty"`
	// Human-readable message indicating details about last transition.
	Message string `json:"message,omitempty"`
}

// ConditionType is a valid value for Conidtion.Type
type ConditionType string

const (
	// Reconciled indicates whether the resource was successfully reconciled
	Reconciled ConditionType = "Reconciled"
	// Ready indicates whether the resource is considered ready
	Ready ConditionType = "Ready"
)

//func init() {
//	SchemeBuilder.Register(&Status{})
//}
