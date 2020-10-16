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

package v1alpha1

import (
	scalewaymetav1alpha1 "github.com/scaleway/scaleway-operator/apis/meta/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// RDBUserSpec defines the desired state of RDBUser
type RDBUserSpec struct {
	// UserName is the user name to be created on the RDBInstance
	UserName string `json:"userName"`
	// Password is the password associated to the user
	Password RDBInstancePassword `json:"password"`
	// Admin represents whether the user is an admin user
	// +kubebuilder:default:true
	// +optional
	Admin bool `json:"admin,omitempty"`
	// Privileges represents the privileges given to this user
	Privileges []RDBPrivilege `json:"privileges,omitempty"`
	// InstanceRef represents the reference to the instance of the user
	InstanceRef RDBInstanceRef `json:"instanceRef"`
}

// RDBPrivilege defines a privilege linked to a RDBUser
type RDBPrivilege struct {
	// DatabaseName is the name to a RDB Database for this privilege
	DatabaseName string `json:"databaseRef"`
	// Permission is the given permission for this privilege
	Permission RDBPermission `json:"permission"`
}

// RDBPermission defines a permission for a privilege
// +kubebuilder:validation:Enum=ReadOnly;ReadWrite;All;None
type RDBPermission string

const (
	// PermissionReadOnly is the readonly permission
	PermissionReadOnly RDBPermission = "ReadOnly"
	// PermissionReadWrite is the readwrite permission
	PermissionReadWrite RDBPermission = "ReadWrite"
	// PermissionAll is the all permission
	PermissionAll RDBPermission = "All"
	// PermissionNone is the none permission
	PermissionNone RDBPermission = "None"
)

// RDBInstancePassword defines the password of a RDBInstance
// One of Value or ValueFrom must be specified
type RDBInstancePassword struct {
	// Value represents a raw value
	// +optional
	Value *string `json:"value,omitempty"`
	// ValueFrom represents a value from a secret
	// +optional
	ValueFrom *RDBInstancePasswordValueFrom `json:"valueFrom,omitempty"`
}

// RDBInstancePasswordValueFrom defines a source to get a password from
type RDBInstancePasswordValueFrom struct {
	SecretKeyRef corev1.SecretReference `json:"secretKeyRef"`
}

// RDBUserStatus defines the observed state of RDBUser
type RDBUserStatus struct {
	// Conditions is the current conditions of the RDBInstance
	scalewaymetav1alpha1.Status `json:",inline"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=rdbu;rdbuser
// +kubebuilder:printcolumn:name="UserName",type="string",JSONPath=".spec.userName"

// RDBUser is the Schema for the rdbusers API
type RDBUser struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RDBUserSpec   `json:"spec,omitempty"`
	Status RDBUserStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// RDBUserList contains a list of RDBUser
type RDBUserList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []RDBUser `json:"items"`
}

func init() {
	SchemeBuilder.Register(&RDBUser{}, &RDBUserList{})
}
