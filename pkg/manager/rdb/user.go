package rdb

import (
	"context"
	"fmt"

	"github.com/scaleway/scaleway-operator/pkg/manager/scaleway"
	"github.com/scaleway/scaleway-sdk-go/api/rdb/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	rdbv1alpha1 "github.com/scaleway/scaleway-operator/apis/rdb/v1alpha1"
)

const (
	// SecretPasswordKey is the key for accessing the pasword in the given secret
	SecretPasswordKey = "password"
)

// UserManager manages the RDB users
type UserManager struct {
	client.Client
	API *rdb.API
	scaleway.Manager
}

// Ensure reconciles the RDB user resource
func (m *UserManager) Ensure(ctx context.Context, obj runtime.Object) (bool, error) {
	user, err := convertUser(obj)
	if err != nil {
		return false, err
	}

	instanceID, region, err := m.getInstanceIDAndRegion(ctx, user)
	if err != nil {
		return false, err
	}

	rdbUser, err := m.getByName(ctx, user)
	if err != nil {
		return false, err
	}

	if rdbUser != nil {
		if rdbUser.IsAdmin != user.Spec.Admin {
			_, err := m.API.UpdateUser(&rdb.UpdateUserRequest{
				Region:     region,
				InstanceID: instanceID,
				Name:       rdbUser.Name,
				IsAdmin:    scw.BoolPtr(user.Spec.Admin),
			})
			if err != nil {
				return false, err
			}
		}
		// pass for now if password changed
	} else {
		password := ""

		if user.Spec.Password.ValueFrom != nil {
			secret := corev1.Secret{}
			err := m.Get(ctx, types.NamespacedName{
				Name:      user.Spec.Password.ValueFrom.SecretKeyRef.Name,
				Namespace: user.Spec.Password.ValueFrom.SecretKeyRef.Namespace,
			}, &secret)
			if err != nil {
				return false, err
			}
			password = string(secret.Data[SecretPasswordKey])
		} else if user.Spec.Password.Value != nil {
			password = *user.Spec.Password.Value
		}

		rdbUser, err = m.API.CreateUser(&rdb.CreateUserRequest{
			Region:     region,
			InstanceID: instanceID,
			IsAdmin:    user.Spec.Admin,
			Name:       user.Spec.UserName,
			Password:   password,
		})
		if err != nil {
			return false, err
		}
	}

	return false, nil
}

// Delete deletes the RDB user resource
func (m *UserManager) Delete(ctx context.Context, obj runtime.Object) (bool, error) {
	user, err := convertUser(obj)
	if err != nil {
		return false, err
	}

	instanceID, region, err := m.getInstanceIDAndRegion(ctx, user)
	if err != nil {
		return false, err
	}

	if instanceID == "" {
		return true, nil
	}

	err = m.API.DeleteUser(&rdb.DeleteUserRequest{
		Region:     region,
		InstanceID: instanceID,
		Name:       user.Spec.UserName,
	})
	if err != nil {
		if _, ok := err.(*scw.ResourceNotFoundError); ok {
			return true, nil
		}
		return false, err
	}

	return false, nil
}

// GetOwners returns the owners of the RDB user resource
func (m *UserManager) GetOwners(ctx context.Context, obj runtime.Object) ([]scaleway.Owner, error) {
	user, err := convertUser(obj)
	if err != nil {
		return nil, err
	}

	if user.Spec.InstanceRef.Name == "" {
		return nil, nil
	}

	instance := &rdbv1alpha1.RDBInstance{}

	userNamespace := user.Spec.InstanceRef.Namespace
	if userNamespace == "" {
		userNamespace = user.Namespace
	}

	err = m.Get(ctx, client.ObjectKey{Name: user.Spec.InstanceRef.Name, Namespace: userNamespace}, instance)
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

func (m *UserManager) getInstanceIDAndRegion(ctx context.Context, user *rdbv1alpha1.RDBUser) (string, scw.Region, error) {
	instance := &rdbv1alpha1.RDBInstance{}

	if user.Spec.InstanceRef.Name != "" {
		userNamespace := user.Spec.InstanceRef.Namespace
		if userNamespace == "" {
			userNamespace = user.Namespace
		}

		err := m.Get(ctx, client.ObjectKey{Name: user.Spec.InstanceRef.Name, Namespace: userNamespace}, instance)
		if err != nil {
			return "", "", err
		}

		return instance.Spec.InstanceID, scw.Region(instance.Spec.Region), nil
	}

	return user.Spec.InstanceRef.ExternalID, scw.Region(user.Spec.InstanceRef.Region), nil
}

func (m *UserManager) getByName(ctx context.Context, user *rdbv1alpha1.RDBUser) (*rdb.User, error) {
	instanceID, region, err := m.getInstanceIDAndRegion(ctx, user)
	if err != nil {
		return nil, err
	}

	usersResp, err := m.API.ListUsers(&rdb.ListUsersRequest{
		Region:     region,
		InstanceID: instanceID,
		Name:       scw.StringPtr(user.Spec.UserName),
	}, scw.WithAllPages())
	if err != nil {
		if _, ok := err.(*scw.ResourceNotFoundError); ok {
			return nil, nil
		}
		return nil, err
	}

	var finalUser *rdb.User

	for _, rdbUser := range usersResp.Users {
		if rdbUser.Name == user.Spec.UserName {
			finalUser = rdbUser
			break
		}
	}

	if finalUser != nil {
		return finalUser, nil
	}

	return nil, nil
}

func convertUser(obj runtime.Object) (*rdbv1alpha1.RDBUser, error) {
	user, ok := obj.(*rdbv1alpha1.RDBUser)
	if !ok {
		return nil, fmt.Errorf("failed type assertion on kind: %s", obj.GetObjectKind().GroupVersionKind().String())
	}
	return user, nil
}
