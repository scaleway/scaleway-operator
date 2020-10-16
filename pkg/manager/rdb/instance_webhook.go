package rdb

import (
	"context"

	"github.com/scaleway/scaleway-sdk-go/api/rdb/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

// ValidateCreate validates the creation of a RDB Instance
func (m *InstanceManager) ValidateCreate(ctx context.Context, obj runtime.Object) (field.ErrorList, error) {
	var allErrs field.ErrorList

	instance, err := convertInstance(obj)
	if err != nil {
		return nil, err
	}
	_, err = scw.ParseRegion(instance.Spec.Region)
	if instance.Spec.Region != "" && err != nil {
		allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("region"), instance.Spec.Region, "region is not valid"))
		return allErrs, nil // stop validation here since future calls will fail
	}

	enginesResp, err := m.API.ListDatabaseEngines(&rdb.ListDatabaseEnginesRequest{
		Region: scw.Region(instance.Spec.Region),
	}, scw.WithAllPages())
	if err != nil {
		return nil, err
	}

	if instance.Spec.InstanceID != "" {
		rdbInstance, err := m.API.GetInstance(&rdb.GetInstanceRequest{
			Region:     scw.Region(instance.Spec.Region),
			InstanceID: instance.Spec.InstanceID,
		})
		if err != nil {
			allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("instanceID"), instance.Spec.InstanceID, err.Error()))
			return allErrs, nil
		}
		if instance.Spec.Engine != rdbInstance.Engine {
			allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("engine"), instance.Spec.Engine, "engine does not match"))
		}

		return allErrs, nil
	}

	engineFound := false
	for _, engine := range enginesResp.Engines {
		for _, engineVersion := range engine.Versions {
			if engineVersion.Name == instance.Spec.Engine {
				if engineVersion.Disabled {
					allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("engine"), instance.Spec.Engine, "engine is disabled"))
				}
				engineFound = true
				break
			}
		}
	}
	if !engineFound {
		allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("engine"), instance.Spec.Engine, "engine does not exist"))
	}

	nodeTypeErrs, err := m.checkNodeType(ctx, scw.Region(instance.Spec.Region), instance.Spec.NodeType)
	if err != nil {
		return nil, err
	}
	allErrs = append(allErrs, nodeTypeErrs...)

	return allErrs, nil
}

// ValidateUpdate validates the update of a RDB Instance
func (m *InstanceManager) ValidateUpdate(ctx context.Context, oldObj runtime.Object, obj runtime.Object) (field.ErrorList, error) {
	var allErrs field.ErrorList

	instance, err := convertInstance(obj)
	if err != nil {
		return nil, err
	}

	oldInstance, err := convertInstance(oldObj)
	if err != nil {
		return nil, err
	}

	if oldInstance.Spec.InstanceID != "" && oldInstance.Spec.InstanceID != instance.Spec.InstanceID {
		allErrs = append(allErrs, field.Forbidden(field.NewPath("spec").Child("instanceID"), "field is immutable"))
	}

	if oldInstance.Spec.Region != "" && oldInstance.Spec.Region != instance.Spec.Region {
		allErrs = append(allErrs, field.Forbidden(field.NewPath("spec").Child("region"), "field is immutable"))
	}

	if oldInstance.Spec.Engine != instance.Spec.Engine {
		allErrs = append(allErrs, field.Forbidden(field.NewPath("spec").Child("engine"), "field is immutable"))
	}

	if oldInstance.Spec.IsHaCluster != instance.Spec.IsHaCluster && oldInstance.Spec.IsHaCluster {
		allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("isHaCluster"), instance.Spec.Engine, "HA instance can't be downgraded"))
	}

	if oldInstance.Spec.NodeType != instance.Spec.NodeType {
		nodeTypeErrs, err := m.checkNodeType(ctx, scw.Region(instance.Spec.Region), instance.Spec.NodeType)
		if err != nil {
			return nil, err
		}
		allErrs = append(allErrs, nodeTypeErrs...)
	}

	return allErrs, nil
}

func (m *InstanceManager) checkNodeType(ctx context.Context, region scw.Region, instanceNodeType string) (field.ErrorList, error) {
	var allErrs field.ErrorList

	nodeTypesResp, err := m.API.ListNodeTypes(&rdb.ListNodeTypesRequest{
		Region: scw.Region(region),
	}, scw.WithAllPages())
	if err != nil {
		return nil, err
	}

	nodeTypeFound := false
	for _, nodeType := range nodeTypesResp.NodeTypes {
		if nodeType.Name == instanceNodeType {
			if nodeType.Disabled {
				allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("nodeType"), instanceNodeType, "node type is disabled"))
			}
			nodeTypeFound = true
			break
		}
	}
	if !nodeTypeFound {
		allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("nodeType"), instanceNodeType, "node type does not exist"))
	}

	return allErrs, nil
}
