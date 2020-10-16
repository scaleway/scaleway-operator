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
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// RDBDatabaseSpec defines the desired state of RDBDatabase
type RDBDatabaseSpec struct {
	// InstanceRef represents the reference to the instance of the database
	InstanceRef RDBInstanceRef `json:"instanceRef"`
	// OverrideName represents the name given to the database
	// This field is immutable after creation
	// +optional
	OverrideName string `json:"overrideName,omitempty"`
}

// RDBInstanceRef defines a reference to rdb instance
// Only one of ExternalID/Region or Name/Namespace must be specified
type RDBInstanceRef struct {
	// ExternalID is the ID of the instance
	// This field is immutable after creation
	// +optional
	ExternalID string `json:"externalID,omitempty"`
	// Region is the region of the instance
	// This field is immutable after creation
	// +optional
	Region string `json:"region,omitempty"`
	// Name is the name of the instance of this database
	// This field is immutable after creation
	// +optional
	Name string `json:"name,omitempty"`
	// Namespace is the namespace of the instance of this database
	// If empty, it will use the namespace of the database
	// This field is immutable after creation
	// +optional
	Namespace string `json:"namespace,omitempty"`
}

// RDBDatabaseStatus defines the observed state of RDBDatabase
type RDBDatabaseStatus struct {
	// Size represents the size of the database
	Size *resource.Quantity `json:"size,omitempty"`
	// Managed defines whether this database is mananged
	Managed bool `json:"managed,omitempty"`
	// Owner represents the owner of this database
	Owner string `json:"owner,omitempty"`
	// Conditions is the current conditions of the RDBDatabase
	scalewaymetav1alpha1.Status `json:",inline"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=rdbd;rdbdatabase
// +kubebuilder:printcolumn:name="size",type="string",JSONPath=".status.size"
// +kubebuilder:printcolumn:name="owner",type="string",JSONPath=".status.owner"
// +kubebuilder:printcolumn:name="managed",type="string",JSONPath=".status.managed"

// RDBDatabase is the Schema for the rdbdatabases API
type RDBDatabase struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RDBDatabaseSpec   `json:"spec,omitempty"`
	Status RDBDatabaseStatus `json:"status,omitempty"`
}

// GetStatus returns the scaleway meta status
func (r *RDBDatabase) GetStatus() scalewaymetav1alpha1.Status {
	return r.Status.Status
}

// SetStatus sets the scaleway meta status
func (r *RDBDatabase) SetStatus(status scalewaymetav1alpha1.Status) {
	r.Status.Status = status
}

// +kubebuilder:object:root=true

// RDBDatabaseList contains a list of RDBDatabase
type RDBDatabaseList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []RDBDatabase `json:"items"`
}

func init() {
	SchemeBuilder.Register(&RDBDatabase{}, &RDBDatabaseList{})
}
