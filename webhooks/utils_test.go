package webhooks

import (
	"testing"

	"k8s.io/apimachinery/pkg/runtime/schema"
)

func Test_GenerateValidatePath(t *testing.T) {
	cases := []struct {
		gvk    schema.GroupVersionKind
		output string
	}{
		{
			schema.GroupVersionKind{
				Group:   "foo",
				Version: "v2",
				Kind:    "bar",
			},
			"/validate-foo-v2-bar",
		},
		{
			schema.GroupVersionKind{
				Group:   "my-awesome-group",
				Version: "v99",
				Kind:    "MyNiceKind",
			},
			"/validate-my-awesome-group-v99-mynicekind",
		},
	}

	for _, c := range cases {
		output := GenerateValidatePath(c.gvk)
		if output != c.output {
			t.Errorf("Got %s instead of %s", output, c.output)
		}
	}
}
