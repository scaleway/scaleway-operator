package rdb

import (
	"context"

	"github.com/scaleway/scaleway-sdk-go/api/rdb/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

// ValidateCreate validates the creation of a RDB Database
func (m *DatabaseManager) ValidateCreate(ctx context.Context, obj runtime.Object) (field.ErrorList, error) {
	var allErrs field.ErrorList

	database, err := convertDatabase(obj)
	if err != nil {
		return nil, err
	}

	byName := database.Spec.InstanceRef.Name != "" || database.Spec.InstanceRef.Namespace != ""
	byID := database.Spec.InstanceRef.ExternalID != "" || database.Spec.InstanceRef.Region != ""

	if database.Spec.InstanceRef.Name == "" && database.Spec.InstanceRef.ExternalID == "" {
		allErrs = append(allErrs, field.Required(field.NewPath("spec").Child("instanceRef"), "name/namespace or externalID/region must be specified"))
		return allErrs, nil
	}
	if byName && byID {
		allErrs = append(allErrs, field.Forbidden(field.NewPath("spec").Child("instanceRef"), "only on of name/namespace and externalID/region must be specified"))
		return allErrs, nil
	}

	if byID {
		_, err = scw.ParseRegion(database.Spec.InstanceRef.Region)
		if database.Spec.InstanceRef.Region != "" && err != nil {
			allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("instanceRef").Child("region"), database.Spec.InstanceRef.Region, "region is not valid"))
			return allErrs, nil
		}

		_, err = m.API.GetInstance(&rdb.GetInstanceRequest{
			Region:     scw.Region(database.Spec.InstanceRef.Region),
			InstanceID: database.Spec.InstanceRef.ExternalID,
		})
		if err != nil {
			allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("instanceRef").Child("externalID"), database.Spec.InstanceRef.ExternalID, err.Error()))
			return allErrs, nil
		}
	}

	return allErrs, nil
}

// ValidateUpdate validates the update of a RDB Database
func (m *DatabaseManager) ValidateUpdate(ctx context.Context, oldObj runtime.Object, obj runtime.Object) (field.ErrorList, error) {
	var allErrs field.ErrorList

	database, err := convertDatabase(obj)
	if err != nil {
		return nil, err
	}

	oldDatabase, err := convertDatabase(oldObj)
	if err != nil {
		return nil, err
	}

	if oldDatabase.Spec.OverrideName != oldDatabase.Spec.OverrideName {
		allErrs = append(allErrs, field.Forbidden(field.NewPath("spec").Child("overrideName"), "field is immutable"))
	}

	if oldDatabase.Spec.InstanceRef.Region != database.Spec.InstanceRef.Region {
		allErrs = append(allErrs, field.Forbidden(field.NewPath("spec").Child("instanceRef").Child("region"), "field is immutable"))
	}

	if oldDatabase.Spec.InstanceRef.ExternalID != database.Spec.InstanceRef.ExternalID {
		allErrs = append(allErrs, field.Forbidden(field.NewPath("spec").Child("instanceRef").Child("externalID"), "field is immutable"))
	}

	if oldDatabase.Spec.InstanceRef.Name != database.Spec.InstanceRef.Name {
		allErrs = append(allErrs, field.Forbidden(field.NewPath("spec").Child("instanceRef").Child("name"), "field is immutable"))
	}

	if oldDatabase.Spec.InstanceRef.Namespace != database.Spec.InstanceRef.Namespace {
		allErrs = append(allErrs, field.Forbidden(field.NewPath("spec").Child("instanceRef").Child("namespace"), "field is immutable"))
	}

	return allErrs, nil
}
