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

// CloudflareAccessReusablePolicySpec defines the desired state of CloudflareAccessReusablePolicy.
type CloudflareAccessReusablePolicySpec struct {
	// Name of the Cloudflare Access's reusable Policy
	Name string `json:"name"`

	// The action Access will take if a user matches this policy. Infrastructure application policies can only use the Allow action.
	// +kubebuilder:validation:Enum=allow;deny;non_identity;bypass
	// +kubebuilder:default=allow
	Decision string `json:"decision"`

	// Rules evaluated with an OR logical operator. A user needs to meet only one of the Include rules.
	Include []CloudFlareAccessRule `json:"include"`

	// Rules evaluated with an AND logical operator. To match the policy, a user must meet all of the Require rules.
	// +optional
	Require []CloudFlareAccessRule `json:"require,omitempty"`

	// Rules evaluated with a NOT logical operator. To match the policy, a user cannot meet any of the Exclude rules.
	// +optional
	Exclude []CloudFlareAccessRule `json:"exclude,omitempty"`
}

func (c CloudflareAccessReusablePolicySpec) GetInclude() []CloudFlareAccessRule {
	return c.Include
}

func (c CloudflareAccessReusablePolicySpec) GetExclude() []CloudFlareAccessRule {
	return c.Exclude
}

func (c CloudflareAccessReusablePolicySpec) GetRequire() []CloudFlareAccessRule {
	return c.Require
}

// CloudflareAccessReusablePolicyStatus defines the observed state of CloudflareAccessReusablePolicy.
type CloudflareAccessReusablePolicyStatus struct {
	// AccessReusablePolicyID is the ID of the reference in Cloudflare
	AccessReusablePolicyID string `json:"accessReusablePolicyId,omitempty"`

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

// CloudflareAccessReusablePolicy is the Schema for the cloudflareaccessreusablepolicies API.
type CloudflareAccessReusablePolicy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`

	Spec   CloudflareAccessReusablePolicySpec   `json:"spec"`
	Status CloudflareAccessReusablePolicyStatus `json:"status"`
}

func (c *CloudflareAccessReusablePolicy) GetType() string {
	return "CloudflareAccessReusablePolicy"
}

func (c *CloudflareAccessReusablePolicy) GetID() string {
	return c.Status.AccessReusablePolicyID
}

func (c *CloudflareAccessReusablePolicy) UnderDeletion() bool {
	return !c.DeletionTimestamp.IsZero()
}

// +kubebuilder:object:root=true

// CloudflareAccessReusablePolicyList contains a list of CloudflareAccessReusablePolicy.
type CloudflareAccessReusablePolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []CloudflareAccessReusablePolicy `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CloudflareAccessReusablePolicy{}, &CloudflareAccessReusablePolicyList{})
}
