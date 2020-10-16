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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// RDBInstanceSpec defines the desired state of RDBInstance
type RDBInstanceSpec struct {
	// InstanceID is the ID of the instance
	// If empty it will create a new instance
	// If set it will use this ID as the instance ID
	// This field is immutable after creation
	// At most one of InstanceID/Region and InstanceFrom have to be specified
	// on creation.
	// +optional
	InstanceID string `json:"instanceID,omitempty"`
	// Region is the region in which the RDBInstance will run
	// This field is immutable after creation
	// Defaults to the controller default region
	// At most one of InstanceID/Region and InstanceFrom have to be specified
	// on creation.
	// +optional
	Region string `json:"region,omitempty"`
	// InstanceFrom allows to create an instance from an existing one
	// At most one of InstanceID/Region and InstanceFrom have to be specified
	// on creation.
	// This field is immutable after creation
	// +optional
	InstanceFrom *RDBInstanceRef `json:"instanceFrom,omitempty"`
	// Engine is the database engine of the RDBInstance
	Engine string `json:"engine"`
	// NodeType is the type of node to use for the RDBInstance
	NodeType string `json:"nodeType"`
	// IsHaCluster represents whether the RDBInstance should be in HA mode
	// Defaults to false
	// +kubebuilder:default:false
	// +optional
	IsHaCluster bool `json:"isHaCluster,omitempty"`
	// AutoBackup represents the RDBInstance auto backup policy
	// +optional
	AutoBackup *RDBInstanceAutoBackup `json:"autoBackup,omitempty"`
	// ACL represents the ACL rules of the RDBInstance
	ACL *RDBACL `json:"acl,omitempty"`
}

// RDBACL defines the acl of a RDBInstance
type RDBACL struct {
	// Rules represents the RDB ACL rules
	// +optional
	Rules []RDBACLRule `json:"rules,omitempty"`

	// AllowCluster represents wether the nodes in the cluster
	// should be allowed
	// +optional
	AllowCluster bool `json:"allowCluster,omitempty"`
}

// RDBACLRule defines a rule for a RDB ACL
type RDBACLRule struct {
	// IPRange represents a CIDR IP range
	IPRange string `json:"ipRange"`
	// Description is the description associated with this ACL rule
	// +optional
	Description string `json:"description,omitempty"`
}

// RDBInstanceAutoBackup defines the auto backup state of a RDBInstance
type RDBInstanceAutoBackup struct {
	// Disabled represents whether the auto backup should be disabled
	Disabled bool `json:"disabled,omitempty"`
	// Frequency represents the frequency, in hour, at which auto backups are made
	// Default to 24
	// +kubebuilder:default:24
	// +kubebuilder:validation:Minimum=0
	// +optional
	Frequency *int32 `json:"frequency,omitempty"`
	// Retention represents the number of days the autobackup are kept
	// Default to 7
	// +kubebuilder:default:7
	// +kubebuilder:validation:Minimum=0
	// +optional
	Retention *int32 `json:"retention,omitempty"`
}

// RDBInstanceStatus defines the observed state of RDBInstance
type RDBInstanceStatus struct {
	// Endpoint is the endpoint of the RDBInstance
	Endpoint RDBInstanceEndpoint `json:"endpoint,omitempty"`
	// Conditions is the current conditions of the RDBInstance
	scalewaymetav1alpha1.Status `json:",inline"`
}

// RDBInstanceEndpoint defines the endpoint of a RDBInstance
type RDBInstanceEndpoint struct {
	// IP is the IP of the RDBInstance
	IP string `json:"ip,omitempty"`
	// Port if the port of the RDBInstance
	Port int32 `json:"port,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=rdbi;rdbinstance
// +kubebuilder:printcolumn:name="IP",type="string",JSONPath=".status.endpoint.ip"
// +kubebuilder:printcolumn:name="Port",type="integer",JSONPath=".status.endpoint.port"

// RDBInstance is the Schema for the databaseinstances API
type RDBInstance struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RDBInstanceSpec   `json:"spec,omitempty"`
	Status RDBInstanceStatus `json:"status,omitempty"`
}

// GetStatus returns the scaleway meta status
func (r *RDBInstance) GetStatus() scalewaymetav1alpha1.Status {
	return r.Status.Status
}

// SetStatus sets the scaleway meta status
func (r *RDBInstance) SetStatus(status scalewaymetav1alpha1.Status) {
	r.Status.Status = status
}

// +kubebuilder:object:root=true

// RDBInstanceList contains a list of RDBInstance
type RDBInstanceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []RDBInstance `json:"items"`
}

func init() {
	SchemeBuilder.Register(&RDBInstance{}, &RDBInstanceList{})
}
