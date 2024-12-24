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
	"github.com/bojanzelic/cloudflare-zero-trust-operator/internal/cfapi"
	cloudflare "github.com/cloudflare/cloudflare-go"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// CloudflareAccessGroupSpec defines the desired state of CloudflareAccessGroup.
type CloudflareAccessGroupSpec struct {
	// Name of the Cloudflare Access Group
	Name string `json:"name"`

	// Rules evaluated with an OR logical operator. A user needs to meet only one of the Include rules.
	Include []CloudFlareAccessGroupRule `json:"include,omitempty"`

	// Rules evaluated with an AND logical operator. To match the policy, a user must meet all of the Require rules.
	Require []CloudFlareAccessGroupRule `json:"require,omitempty"`

	// Rules evaluated with a NOT logical operator. To match the policy, a user cannot meet any of the Exclude rules.
	Exclude []CloudFlareAccessGroupRule `json:"exclude,omitempty"`
}

type CloudFlareAccessGroupRule struct {
	// Matches a Specific email
	Emails []string `json:"emails,omitempty"`

	// Matches a specific email Domain
	EmailDomains []string `json:"emailDomains,omitempty"`

	// Matches an IP CIDR block
	IPRanges []string `json:"ipRanges,omitempty"`

	// Reference to other access groups
	AccessGroups []AccessGroup `json:"accessGroups,omitempty"`

	// Country
	Country []string `json:"country,omitempty"`

	// Allow Everyone
	Everyone *bool `json:"everyone,omitempty"`

	// Certificate CN
	CommonName []string `json:"commonName,omitempty"`

	// Any valid certificate will be matched
	ValidCertificate *bool `json:"validCertificate,omitempty"`

	// Matches a service token
	ServiceToken []ServiceToken `json:"serviceToken,omitempty"`

	// Matches any valid service token
	AnyAccessServiceToken *bool `json:"anyAccessServiceToken,omitempty"`

	// ID of the login method
	LoginMethod []string `json:"loginMethod,omitempty"`

	// Matches Google Group
	GoogleGroups []GoogleGroup `json:"googleGroups,omitempty"`

	// Okta Groups
	OktaGroup []OktaGroup `json:"oktaGroup,omitempty"`

	// OIDC Claims
	OIDCClaims []OIDCClaim `json:"oidcClaims,omitempty"`
}

// CloudflareAccessGroupStatus defines the observed state of CloudflareAccessGroup.
type CloudflareAccessGroupStatus struct {
	// AccessGroupID is the ID of the reference in Cloudflare
	AccessGroupID string `json:"accessGroupId,omitempty"`

	// Creation timestamp of the resource in Cloudflare
	CreatedAt metav1.Time `json:"createdAt,omitempty"`

	// Updated timestamp of the resource in Cloudflare
	UpdatedAt metav1.Time `json:"updatedAt,omitempty"`

	// Conditions store the status conditions of the CloudflareAccessApplication
	// +operator-sdk:csv:customresourcedefinitions:type=status
	Conditions []metav1.Condition `json:"conditions,omitempty" patchMergeKey:"type" patchStrategy:"merge" protobuf:"bytes,1,rep,name=conditions"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// CloudflareAccessGroup is the Schema for the cloudflareaccessgroups API.
type CloudflareAccessGroup struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CloudflareAccessGroupSpec   `json:"spec,omitempty"`
	Status CloudflareAccessGroupStatus `json:"status,omitempty"`
}

func (c *CloudflareAccessGroup) GetType() string {
	return "CloudflareAccessGroup"
}

func (c *CloudflareAccessGroup) GetID() string {
	return c.Status.AccessGroupID
}

func (c *CloudflareAccessGroup) UnderDeletion() bool {
	return !c.ObjectMeta.DeletionTimestamp.IsZero()
}

func (c *CloudflareAccessGroup) ToCloudflare() cloudflare.AccessGroup {
	accessGroup := cloudflare.AccessGroup{
		Name:      c.Spec.Name,
		ID:        c.Status.AccessGroupID,
		CreatedAt: &c.Status.CreatedAt.Time,
		UpdatedAt: &c.Status.UpdatedAt.Time,
		Include:   make([]interface{}, 0),
		Exclude:   make([]interface{}, 0),
		Require:   make([]interface{}, 0),
	}

	managedCRFields := CloudFlareAccessGroupRuleGroups{
		c.Spec.Include,
		c.Spec.Exclude,
		c.Spec.Require,
	}

	managedCFFields := []*[]interface{}{
		&accessGroup.Include,
		&accessGroup.Exclude,
		&accessGroup.Require,
	}

	managedCRFields.TransformCloudflareRuleFields(managedCFFields)

	return accessGroup
}

type CloudFlareAccessGroupRuleGroups [][]CloudFlareAccessGroupRule

// nolint: gocognit,cyclop
func (c CloudFlareAccessGroupRuleGroups) TransformCloudflareRuleFields(managedCFFields []*[]interface{}) {
	for i, managedField := range c {
		for _, field := range managedField {
			for _, email := range field.Emails {
				*managedCFFields[i] = append(*managedCFFields[i], cfapi.NewAccessGroupEmail(email))
			}
			for _, domain := range field.EmailDomains {
				*managedCFFields[i] = append(*managedCFFields[i], cfapi.NewAccessGroupEmailDomains(domain))
			}
			for _, ip := range field.IPRanges {
				*managedCFFields[i] = append(*managedCFFields[i], cfapi.NewAccessGroupIP(ip))
			}
			for _, token := range field.ServiceToken {
				if token.Value != "" {
					*managedCFFields[i] = append(*managedCFFields[i], cfapi.NewAccessGroupServiceToken(token.Value))
				}
			}
			if field.AnyAccessServiceToken != nil && *field.AnyAccessServiceToken {
				*managedCFFields[i] = append(*managedCFFields[i], cfapi.NewAccessGroupAnyValidServiceToken())
			}
			if field.Everyone != nil && *field.Everyone {
				*managedCFFields[i] = append(*managedCFFields[i], cfapi.NewAccessGroupEveryone())
			}
			if field.ValidCertificate != nil && *field.ValidCertificate {
				*managedCFFields[i] = append(*managedCFFields[i], cfapi.NewAccessGroupCertificate())
			}
			for _, country := range field.Country {
				*managedCFFields[i] = append(*managedCFFields[i], cfapi.NewAccessGroupGeo(country))
			}
			for _, group := range field.AccessGroups {
				if group.Value != "" {
					*managedCFFields[i] = append(*managedCFFields[i], cfapi.NewAccessGroupAccessGroup(group.Value))
				}
			}

			for _, loginMethod := range field.LoginMethod {
				*managedCFFields[i] = append(*managedCFFields[i], cfapi.NewAccessGroupLoginMethod(loginMethod))
			}

			for _, googleGroup := range field.GoogleGroups {
				if googleGroup.Email != "" && googleGroup.IdentityProviderID != "" {
					*managedCFFields[i] = append(*managedCFFields[i], cfapi.NewAccessGroupGSuite(googleGroup.Email, googleGroup.IdentityProviderID))
				}
			}

			for _, oktaGroup := range field.OktaGroup {
				*managedCFFields[i] = append(*managedCFFields[i], cfapi.NewAccessGroupOktaGroup(oktaGroup.Name, oktaGroup.IdentityProviderID))
			}

			for _, oidcClaim := range field.OIDCClaims {
				*managedCFFields[i] = append(*managedCFFields[i], cfapi.NewAccessGroupOIDCClaim(oidcClaim.Name, oidcClaim.Value, oidcClaim.IdentityProviderID))
			}
		}
	}
}

//+kubebuilder:object:root=true

// CloudflareAccessGroupList contains a list of CloudflareAccessGroup.
type CloudflareAccessGroupList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CloudflareAccessGroup `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CloudflareAccessGroup{}, &CloudflareAccessGroupList{})
}
