/*
Copyright 2025.

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
	Include CloudFlareAccessRules `json:"include"`

	// Rules evaluated with an AND logical operator. To match the policy, a user must meet all the Require rules.
	Require CloudFlareAccessRules `json:"require"`

	// Rules evaluated with a NOT logical operator. To match the policy, a user cannot meet any of the Exclude rules.
	Exclude CloudFlareAccessRules `json:"exclude"`
}

// CloudflareAccessGroupStatus defines the observed state of CloudflareAccessGroup.
type CloudflareAccessGroupStatus struct {
	// AccessGroupID is the ID of the reference in Cloudflare
	AccessGroupID string `json:"accessGroupId,omitzero"`

	// Creation timestamp of the resource in Cloudflare
	CreatedAt metav1.Time `json:"createdAt,omitzero"`

	// Updated timestamp of the resource in Cloudflare
	UpdatedAt metav1.Time `json:"updatedAt,omitzero"`

	//
	ResolvedIdpsFromRefs RulerResolvedCloudflareIDs `json:"resolvedCfIds,omitzero"`

	// Conditions store the status conditions of the CloudflareAccessApplication
	//
	// +operator-sdk:csv:customresourcedefinitions:type=status
	Conditions []metav1.Condition `json:"conditions,omitempty" patchMergeKey:"type" patchStrategy:"merge" protobuf:"bytes,1,rep,name=conditions"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// CloudflareAccessGroup is the Schema for the cloudflareaccessgroups API.
type CloudflareAccessGroup struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitzero"`

	Spec CloudflareAccessGroupSpec `json:"spec,omitzero"`

	// +optional
	Status CloudflareAccessGroupStatus `json:"status,omitzero"`
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

func (c CloudflareAccessGroup) GetIncludeRules() *CloudFlareAccessRules {
	return &c.Spec.Include
}
func (c CloudflareAccessGroup) GetExcludeRules() *CloudFlareAccessRules {
	return &c.Spec.Exclude
}
func (c CloudflareAccessGroup) GetRequireRules() *CloudFlareAccessRules {
	return &c.Spec.Require
}
func (c CloudflareAccessGroup) GetIncludeCfIds() *ResolvedCloudflareIDs {
	return &c.Status.ResolvedIdpsFromRefs.Include
}
func (c CloudflareAccessGroup) GetExcludeCfIds() *ResolvedCloudflareIDs {
	return &c.Status.ResolvedIdpsFromRefs.Exclude
}
func (c CloudflareAccessGroup) GetRequireCfIds() *ResolvedCloudflareIDs {
	return &c.Status.ResolvedIdpsFromRefs.Require
}

// +kubebuilder:object:root=true

// CloudflareAccessGroupList contains a list of CloudflareAccessGroup.
type CloudflareAccessGroupList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitzero"`
	Items           []CloudflareAccessGroup `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CloudflareAccessGroup{}, &CloudflareAccessGroupList{})
}
