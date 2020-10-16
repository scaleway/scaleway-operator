package webhooks

import (
	"strings"

	"k8s.io/apimachinery/pkg/runtime/schema"
)

// GenerateValidatePath returns the path for this validating webhook
func GenerateValidatePath(gvk schema.GroupVersionKind) string {
	return "/validate-" + strings.Replace(gvk.Group, ".", "-", -1) + "-" +
		gvk.Version + "-" + strings.ToLower(gvk.Kind)
}
