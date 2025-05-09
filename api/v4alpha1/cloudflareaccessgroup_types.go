/*
Copyright 2022.

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

package v4alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// CloudflareAccessGroupSpec defines the desired state of CloudflareAccessGroup.
type CloudflareAccessGroupSpec struct {
	// Name of the Cloudflare Access Group
	Name string `json:"name"`

	// Rules evaluated with an OR logical operator. A user needs to meet only one of the Include rules.
	Include CloudFlareAccessRules `json:"include,omitempty"`

	// Rules evaluated with an AND logical operator. To match the policy, a user must meet all the Require rules.
	Require CloudFlareAccessRules `json:"require,omitempty"`

	// Rules evaluated with a NOT logical operator. To match the policy, a user cannot meet any of the Exclude rules.
	Exclude CloudFlareAccessRules `json:"exclude,omitempty"`
}

func (c CloudflareAccessGroupSpec) GetInclude() CloudFlareAccessRules { return c.Include }

func (c CloudflareAccessGroupSpec) GetExclude() CloudFlareAccessRules {
	return c.Exclude
}

func (c CloudflareAccessGroupSpec) GetRequire() CloudFlareAccessRules {
	return c.Require
}

// CloudflareAccessGroupStatus defines the observed state of CloudflareAccessGroup.
type CloudflareAccessGroupStatus struct {
	// AccessGroupID is the ID of the reference in Cloudflare
	AccessGroupID string `json:"accessGroupId,omitempty"`

	// Creation timestamp of the resource in Cloudflare
	CreatedAt metav1.Time `json:"createdAt"`

	// Updated timestamp of the resource in Cloudflare
	UpdatedAt metav1.Time `json:"updatedAt"`

	// Conditions store the status conditions of the CloudflareAccessApplication
	// +operator-sdk:csv:customresourcedefinitions:type=status
	Conditions []metav1.Condition `json:"conditions,omitempty" patchMergeKey:"type" patchStrategy:"merge" protobuf:"bytes,1,rep,name=conditions"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// CloudflareAccessGroup is the Schema for the cloudflareaccessgroups API.
type CloudflareAccessGroup struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`

	Spec   CloudflareAccessGroupSpec   `json:"spec"`
	Status CloudflareAccessGroupStatus `json:"status"`
}

func (c CloudflareAccessGroup) GetType() string {
	return "CloudflareAccessGroup"
}

func (c CloudflareAccessGroup) GetID() string {
	return c.Status.AccessGroupID
}

func (c CloudflareAccessGroup) UnderDeletion() bool {
	return !c.DeletionTimestamp.IsZero()
}

// +kubebuilder:object:root=true

// CloudflareAccessGroupList contains a list of CloudflareAccessGroup.
type CloudflareAccessGroupList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []CloudflareAccessGroup `json:"items"`
}

func (abs *CloudflareAccessGroupList) ToGenericPolicyRuler() []GenericAccessPolicyRuler {
	result := make([]GenericAccessPolicyRuler, 0, len(abs.Items))
	for _, ruler := range abs.Items {
		result = append(result, &ruler.Spec)
	}

	return result
}

func init() {
	SchemeBuilder.Register(&CloudflareAccessGroup{}, &CloudflareAccessGroupList{})
}
