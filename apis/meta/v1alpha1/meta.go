package v1alpha1

// TypeMeta implements the Scaleway Status methods
// +k8s:deepcopy-gen=false
type TypeMeta interface {
	GetStatus() Status
	SetStatus(Status)
}
