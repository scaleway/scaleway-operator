package scaleway

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

// Owner represent an owner of a Scaleway resource
type Owner struct {
	Key    types.NamespacedName
	Object runtime.Object
}

// Manager is the interface implemented by all Scaleway products
type Manager interface {
	// Ensure is the method to implement to reconcile the resource
	Ensure(context.Context, runtime.Object) (bool, error)
	// Delete is the method to implement to delete the resource
	Delete(context.Context, runtime.Object) (bool, error)
	// GetOwners is the method to implement to get the resource's owners
	GetOwners(context.Context, runtime.Object) ([]Owner, error)

	// Webhooks
	// ValidateCreate is the method to implement for the creation validation
	ValidateCreate(context.Context, runtime.Object) (field.ErrorList, error)
	// ValidateUpdate is the method to implement for the update validation
	ValidateUpdate(context.Context, runtime.Object, runtime.Object) (field.ErrorList, error)
}
