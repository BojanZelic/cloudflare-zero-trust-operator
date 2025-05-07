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

package v1alpha1

import (
	"github.com/bojanzelic/cloudflare-zero-trust-operator/internal/cfcollections"
	"github.com/cloudflare/cloudflare-go/v4/zero_trust"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// CloudflareAccessReusablePolicySpec defines the desired state of CloudflareAccessReusablePolicy.
type CloudflareAccessReusablePolicySpec struct {
	// Name of the Cloudflare Access's reusable Policy
	Name string `json:"name"`

	// Decision ex: allow, deny, non_identity, bypass - defaults to allow
	Decision string `json:"decision"`

	// Rules evaluated with an OR logical operator. A user needs to meet only one of the Include rules.
	Include []CloudFlareAccessRule `json:"include,omitempty"`

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
	CreatedAt metav1.Time `json:"createdAt,omitempty"`

	// Updated timestamp of the resource in Cloudflare
	UpdatedAt metav1.Time `json:"updatedAt,omitempty"`

	// Conditions store the status conditions of the CloudflareAccessApplication
	// +operator-sdk:csv:customresourcedefinitions:type=status
	Conditions []metav1.Condition `json:"conditions,omitempty" patchMergeKey:"type" patchStrategy:"merge" protobuf:"bytes,1,rep,name=conditions"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// CloudflareAccessReusablePolicy is the Schema for the cloudflareaccessreusablepolicies API.
type CloudflareAccessReusablePolicy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CloudflareAccessReusablePolicySpec   `json:"spec,omitempty"`
	Status CloudflareAccessReusablePolicyStatus `json:"status,omitempty"`
}

func (c *CloudflareAccessReusablePolicy) GetType() string {
	return "CloudflareAccessReusablePolicy"
}

func (c *CloudflareAccessReusablePolicy) GetID() string {
	return c.Status.AccessReusablePolicyID
}

func (c *CloudflareAccessReusablePolicy) UnderDeletion() bool {
	return !c.ObjectMeta.DeletionTimestamp.IsZero()
}

func (aps CloudflareAccessReusablePolicyList) ToCloudflare() cfcollections.AccessReusablePolicyCollection {
	ret := cfcollections.AccessReusablePolicyCollection{}

	for _, policy := range aps.Items {
		transformed := zero_trust.AccessPolicyListResponse{
			Name:     policy.Name,
			Decision: zero_trust.Decision(policy.Spec.Decision),
			Include:  toAccessRules(&policy.Spec.Include),
			Exclude:  toAccessRules(&policy.Spec.Exclude),
			Require:  toAccessRules(&policy.Spec.Require),
		}

		ret = append(ret, transformed)
	}

	return ret
}

// +kubebuilder:object:root=true

// CloudflareAccessReusablePolicyList contains a list of CloudflareAccessReusablePolicy.
type CloudflareAccessReusablePolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CloudflareAccessReusablePolicy `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CloudflareAccessReusablePolicy{}, &CloudflareAccessReusablePolicyList{})
}
