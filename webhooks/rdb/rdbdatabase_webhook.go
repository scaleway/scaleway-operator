/*


Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package webhooks

import (
	"context"
	"net/http"

	"github.com/go-logr/logr"
	admissionv1beta1 "k8s.io/api/admission/v1beta1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	rdbv1alpha1 "github.com/scaleway/scaleway-operator/apis/rdb/v1alpha1"
	"github.com/scaleway/scaleway-operator/webhooks"
)

// +kubebuilder:webhook:verbs=create;update,path=/validate-rdb-scaleway-com-v1alpha1-rdbdatabase,mutating=false,failurePolicy=fail,groups=rdb.scaleway.com,resources=rdbdatabases,versions=v1alpha1,name=vrdbdatabase.kb.io

// RDBDatabaseValidator is the struct used to validate a RDBDatabase
type RDBDatabaseValidator struct {
	ScalewayWebhook *webhooks.ScalewayWebhook
	*admission.Decoder
	Log logr.Logger
}

// SetupWebhookWithManager registers the RDBDatabase webhook
func (v *RDBDatabaseValidator) SetupWebhookWithManager(mgr ctrl.Manager) error {
	webhookServer := mgr.GetWebhookServer()
	webhookType, err := apiutil.GVKForObject(&rdbv1alpha1.RDBDatabase{}, mgr.GetScheme())
	if err != nil {
		return err
	}
	webhookServer.Register(webhooks.GenerateValidatePath(webhookType), &webhook.Admission{
		Handler: v,
	})
	return nil

}

// Handle handles the main logic of the RDBDatabase webhook
func (v *RDBDatabaseValidator) Handle(ctx context.Context, req admission.Request) admission.Response {
	instance := &rdbv1alpha1.RDBDatabase{}

	err := v.Decode(req, instance)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	var allErrs field.ErrorList

	switch req.Operation {
	case admissionv1beta1.Create:
		allErrs, err = v.ScalewayWebhook.ValidateCreate(ctx, instance)
		if err != nil {
			v.Log.Error(err, "could not validate rdb instance creation")
			return admission.Errored(http.StatusInternalServerError, err)
		}

	case admissionv1beta1.Update:
		oldDatabase := &rdbv1alpha1.RDBDatabase{}
		err = v.DecodeRaw(req.OldObject, oldDatabase)
		if err != nil {
			return admission.Errored(http.StatusBadRequest, err)
		}
		allErrs, err = v.ScalewayWebhook.ValidateUpdate(ctx, oldDatabase, instance)
		if err != nil {
			v.Log.Error(err, "could not validate rdb instance update")
			return admission.Errored(http.StatusInternalServerError, err)
		}
	}

	if len(allErrs) == 0 {
		return admission.Allowed("")
	}

	err = apierrors.NewInvalid(schema.GroupKind{Group: "rdb.scaleway.com", Kind: "RDBDatabase"}, instance.Name, allErrs)

	return admission.Denied(err.Error())
}

// InjectDecoder injects the decoder.
func (v *RDBDatabaseValidator) InjectDecoder(d *admission.Decoder) error {
	v.Decoder = d
	return nil
}
