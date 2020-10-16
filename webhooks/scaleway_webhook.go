package webhooks

import (
	"context"

	"github.com/scaleway/scaleway-operator/pkg/manager/scaleway"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

// ScalewayWebhook is the base webhook for Scaleway products
type ScalewayWebhook struct {
	ScalewayManager scaleway.Manager
}

// ValidateCreate calls the ValidateCreate method of the given resource
func (r *ScalewayWebhook) ValidateCreate(ctx context.Context, obj runtime.Object) (field.ErrorList, error) {
	return r.ScalewayManager.ValidateCreate(ctx, obj)
}

// ValidateUpdate calls the ValidateUpdate method of the given resource
func (r *ScalewayWebhook) ValidateUpdate(ctx context.Context, oldObj runtime.Object, obj runtime.Object) (field.ErrorList, error) {
	return r.ScalewayManager.ValidateUpdate(ctx, oldObj, obj)
}
