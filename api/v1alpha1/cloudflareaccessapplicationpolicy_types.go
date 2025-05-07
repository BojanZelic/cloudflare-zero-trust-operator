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
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

type CloudflareAccessApplicationPolicy struct {
	// Name of the Cloudflare Access's Application specific Policy
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

	// PurposeJustificationRequired *bool                 `json:"purpose_justification_required,omitempty"`
	// PurposeJustificationPrompt   *string               `json:"purpose_justification_prompt,omitempty"`
	// ApprovalRequired             *bool                 `json:"approval_required,omitempty"`
	// ApprovalGroups               []cloudflare.AccessApprovalGroup `json:"approval_groups"`
}

func (c CloudflareAccessApplicationPolicy) GetInclude() []CloudFlareAccessRule {
	return c.Include
}

func (c CloudflareAccessApplicationPolicy) GetExclude() []CloudFlareAccessRule {
	return c.Exclude
}

func (c CloudflareAccessApplicationPolicy) GetRequire() []CloudFlareAccessRule {
	return c.Require
}

type CloudflareAccessApplicationPolicyList []CloudflareAccessApplicationPolicy

func (aps CloudflareAccessApplicationPolicyList) ToCloudflare() cfcollections.AccessApplicationPolicyCollection {
	ret := cfcollections.AccessApplicationPolicyCollection{}

	for i, policy := range aps {
		transformed := zero_trust.AccessApplicationPolicyListResponse{
			Name:       policy.Name,
			Decision:   zero_trust.Decision(policy.Decision),
			Include:    toAccessRules(&policy.Include),
			Exclude:    toAccessRules(&policy.Exclude),
			Require:    toAccessRules(&policy.Require),
			Precedence: int64(i + 1),
		}

		ret = append(ret, transformed)
	}

	return ret
}

func (abs CloudflareAccessApplicationPolicyList) ToGenericPolicyRuler() []GenericAccessPolicyRuler {
	result := make([]GenericAccessPolicyRuler, 0, len(abs))
	for _, ruler := range abs {
		result = append(result, ruler)
	}

	return result
}
