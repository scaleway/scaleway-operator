package rdb

import (
	"context"
	"fmt"

	"github.com/scaleway/scaleway-operator/pkg/manager/scaleway"
	"github.com/scaleway/scaleway-sdk-go/api/rdb/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	rdbv1alpha1 "github.com/scaleway/scaleway-operator/apis/rdb/v1alpha1"
)

// DatabaseManager manages the RDB databsses
type DatabaseManager struct {
	client.Client
	API *rdb.API
	scaleway.Manager
}

// Ensure reconciles the RDB database resource
func (m *DatabaseManager) Ensure(ctx context.Context, obj runtime.Object) (bool, error) {
	database, err := convertDatabase(obj)
	if err != nil {
		return false, err
	}

	instanceID, region, err := m.getInstanceIDAndRegion(ctx, database)
	if err != nil {
		return false, err
	}

	rdbDatabase, err := m.getByName(ctx, database)
	if err != nil {
		return false, err
	}

	if rdbDatabase != nil {
		// pass for now
	} else {
		name := database.GetName()
		if database.Spec.OverrideName != "" {
			name = database.Spec.OverrideName
		}
		rdbDatabase, err = m.API.CreateDatabase(&rdb.CreateDatabaseRequest{
			Region:     region,
			InstanceID: instanceID,
			Name:       name,
		})
		if err != nil {
			return false, err
		}
	}

	database.Status.Managed = rdbDatabase.Managed
	database.Status.Owner = rdbDatabase.Owner
	database.Status.Size = resource.NewQuantity(int64(rdbDatabase.Size), resource.DecimalSI)

	return true, nil
}

// Delete deletes the RDB database resource
func (m *DatabaseManager) Delete(ctx context.Context, obj runtime.Object) (bool, error) {
	database, err := convertDatabase(obj)
	if err != nil {
		return false, err
	}

	instanceID, region, err := m.getInstanceIDAndRegion(ctx, database)
	if err != nil {
		return false, err
	}

	name := database.GetName()
	if database.Spec.OverrideName != "" {
		name = database.Spec.OverrideName
	}

	err = m.API.DeleteDatabase(&rdb.DeleteDatabaseRequest{
		Region:     region,
		InstanceID: instanceID,
		Name:       name,
	})
	if err != nil {
		if _, ok := err.(*scw.ResourceNotFoundError); ok {
			return true, nil
		}
		return false, err
	}

	return true, nil
}

func (m *DatabaseManager) getByName(ctx context.Context, database *rdbv1alpha1.RDBDatabase) (*rdb.Database, error) {
	instanceID, region, err := m.getInstanceIDAndRegion(ctx, database)
	if err != nil {
		return nil, err
	}

	name := database.GetName()
	if database.Spec.OverrideName != "" {
		name = database.Spec.OverrideName
	}

	databasesResp, err := m.API.ListDatabases(&rdb.ListDatabasesRequest{
		Region:     region,
		InstanceID: instanceID,
		Name:       scw.StringPtr(name),
	}, scw.WithAllPages())
	if err != nil {
		if _, ok := err.(*scw.ResourceNotFoundError); ok {
			return nil, nil
		}
		return nil, err
	}

	var finalDatabase *rdb.Database

	for _, rdbDatabase := range databasesResp.Databases {
		if rdbDatabase.Name == name {
			finalDatabase = rdbDatabase
			break
		}
	}

	if finalDatabase != nil {
		return finalDatabase, nil
	}

	return nil, nil
}

// GetOwners returns the owners of the RDB database resource
func (m *DatabaseManager) GetOwners(ctx context.Context, obj runtime.Object) ([]scaleway.Owner, error) {
	database, err := convertDatabase(obj)
	if err != nil {
		return nil, err
	}

	if database.Spec.InstanceRef.Name == "" {
		return nil, nil
	}

	instance := &rdbv1alpha1.RDBInstance{}

	databaseNamespace := database.Spec.InstanceRef.Namespace
	if databaseNamespace == "" {
		databaseNamespace = database.Namespace
	}

	err = m.Get(ctx, client.ObjectKey{Name: database.Spec.InstanceRef.Name, Namespace: databaseNamespace}, instance)
	if err != nil {
		return nil, err
	}

	return []scaleway.Owner{
		{
			Key: types.NamespacedName{
				Name:      instance.Name,
				Namespace: instance.Namespace,
			},
			Object: &rdbv1alpha1.RDBInstance{},
		},
	}, nil
}

func (m *DatabaseManager) getInstanceIDAndRegion(ctx context.Context, database *rdbv1alpha1.RDBDatabase) (string, scw.Region, error) {
	instance := &rdbv1alpha1.RDBInstance{}

	if database.Spec.InstanceRef.Name != "" {
		databaseNamespace := database.Spec.InstanceRef.Namespace
		if databaseNamespace == "" {
			databaseNamespace = database.Namespace
		}

		err := m.Get(ctx, client.ObjectKey{Name: database.Spec.InstanceRef.Name, Namespace: databaseNamespace}, instance)
		if err != nil {
			return "", "", err
		}

		return instance.Spec.InstanceID, scw.Region(instance.Spec.Region), nil
	}

	return database.Spec.InstanceRef.ExternalID, scw.Region(database.Spec.InstanceRef.Region), nil
}

func convertDatabase(obj runtime.Object) (*rdbv1alpha1.RDBDatabase, error) {
	database, ok := obj.(*rdbv1alpha1.RDBDatabase)
	if !ok {
		return nil, fmt.Errorf("failed type assertion on kind: %s", obj.GetObjectKind().GroupVersionKind().String())
	}
	return database, nil
}
