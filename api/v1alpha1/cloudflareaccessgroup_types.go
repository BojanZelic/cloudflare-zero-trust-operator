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
	"github.com/cloudflare/cloudflare-go/v4/zero_trust"
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

func (c CloudflareAccessGroupSpec) GetInclude() []CloudFlareAccessGroupRule {
	return c.Include
}

func (c CloudflareAccessGroupSpec) GetExclude() []CloudFlareAccessGroupRule {
	return c.Exclude
}

func (c CloudflareAccessGroupSpec) GetRequire() []CloudFlareAccessGroupRule {
	return c.Require
}

type CloudFlareAccessGroupRule struct {
	// Matches a Specific email
	Emails []string `json:"emails,omitempty"`

	// Matches a specific email Domain
	EmailDomains []string `json:"emailDomains,omitempty"`

	// Matches an IP CIDR block
	IPRanges []string `json:"ipRanges,omitempty"`

	// Country List
	Countries []string `json:"countries,omitempty"`

	// ID of the login methods
	LoginMethods []string `json:"loginMethods,omitempty"`

	// Certificate CNs
	CommonNames []string `json:"commonNames,omitempty"`

	// Allow Everyone
	Everyone *bool `json:"everyone,omitempty"`

	// Any valid certificate will be matched
	ValidCertificate *bool `json:"validCertificate,omitempty"`

	// Matches any valid service token
	AnyAccessServiceToken *bool `json:"anyAccessServiceToken,omitempty"`

	// Matches a service token
	ServiceTokens []ServiceToken `json:"serviceTokens,omitempty"`

	// Reference to other access groups
	AccessGroups []AccessGroup `json:"accessGroups,omitempty"`

	// Matches Google Group
	GoogleGroups []GoogleGroup `json:"googleGroups,omitempty"`

	// Okta Groups
	OktaGroups []OktaGroup `json:"oktaGroups,omitempty"`

	// OIDC Claims
	OIDCClaims []OIDCClaim `json:"oidcClaims,omitempty"`

	// Github Organizations
	GithubOrganizations []GithubOrganization `json:"githubOrganizations,omitempty"`
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

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

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

func (c *CloudflareAccessGroup) ToCloudflare() zero_trust.AccessGroupGetResponse {
	accessGroup := zero_trust.AccessGroupGetResponse{
		Name:      c.Spec.Name,
		ID:        c.Status.AccessGroupID,
		CreatedAt: c.Status.CreatedAt.Time,
		UpdatedAt: c.Status.UpdatedAt.Time,
		Include:   []zero_trust.AccessRule{},
		Exclude:   []zero_trust.AccessRule{},
		Require:   []zero_trust.AccessRule{},
	}

	managedCRFields := CloudFlareAccessGroupRuleGroups{
		c.Spec.Include,
		c.Spec.Exclude,
		c.Spec.Require,
	}

	managedCFFields := []*[]zero_trust.AccessRule{
		&accessGroup.Include,
		&accessGroup.Exclude,
		&accessGroup.Require,
	}

	managedCRFields.TransformCloudflareRuleFields(managedCFFields)

	return accessGroup
}

type CloudFlareAccessGroupRuleGroups [][]CloudFlareAccessGroupRule

// nolint: gocognit,cyclop
func (c CloudFlareAccessGroupRuleGroups) TransformCloudflareRuleFields(managedCFFields []*[]zero_trust.AccessRule) {
	for i, managedField := range c {
		for _, field := range managedField {
			for _, email := range field.Emails {
				*managedCFFields[i] = append(*managedCFFields[i], zero_trust.AccessRule{
					Email: zero_trust.EmailRuleEmail{
						Email: email,
					},
				})
			}
			for _, domain := range field.EmailDomains {
				*managedCFFields[i] = append(*managedCFFields[i], zero_trust.AccessRule{
					EmailDomain: zero_trust.DomainRuleEmailDomain{
						Domain: domain,
					},
				})
			}
			for _, ip := range field.IPRanges {
				*managedCFFields[i] = append(*managedCFFields[i], zero_trust.AccessRule{
					IP: zero_trust.IPRuleIP{
						IP: ip,
					},
				})
			}
			for _, token := range field.ServiceToken {
				if token.Value != "" {
					*managedCFFields[i] = append(*managedCFFields[i], zero_trust.AccessRule{
						ServiceToken: zero_trust.ServiceTokenRuleServiceToken{
							TokenID: token.Value,
						},
					})
				}
			}
			if field.AnyAccessServiceToken != nil && *field.AnyAccessServiceToken {
				*managedCFFields[i] = append(*managedCFFields[i], zero_trust.AccessRule{
					AnyValidServiceToken: zero_trust.AnyValidServiceTokenRuleAnyValidServiceToken{},
				})
			}
			if field.Everyone != nil && *field.Everyone {
				*managedCFFields[i] = append(*managedCFFields[i], zero_trust.AccessRule{
					Everyone: zero_trust.EveryoneRuleEveryone{},
				})
			}
			if field.ValidCertificate != nil && *field.ValidCertificate {
				*managedCFFields[i] = append(*managedCFFields[i], zero_trust.AccessRule{
					Certificate: zero_trust.CertificateRuleCertificate{},
				})
			}
			for _, countries := range field.Countries {
				*managedCFFields[i] = append(*managedCFFields[i], zero_trust.AccessRule{
					Geo: zero_trust.CountryRuleGeo{
						CountryCode: countries,
					},
				})
			}
			for _, group := range field.AccessGroups {
				if group.Value != "" {
					*managedCFFields[i] = append(*managedCFFields[i], zero_trust.AccessRule{
						Group: zero_trust.GroupRuleGroup{
							ID: group.Value,
						},
					})
				}
			}

			for _, commonNames := range field.CommonNames {
				*managedCFFields[i] = append(*managedCFFields[i], zero_trust.AccessRule{
					CommonNames: zero_trust.AccessRuleAccessCommonNameRuleCommonName{
						CommonNames: commonNames,
					},
				})
			}

			for _, loginMethods := range field.LoginMethods {
				*managedCFFields[i] = append(*managedCFFields[i], zero_trust.AccessRule{
					LoginMethods: zero_trust.AccessRuleAccessLoginMethodRuleLoginMethod{
						ID: loginMethods,
					},
				})
			}

			for _, googleGroup := range field.GoogleGroups {
				if googleGroup.Email != "" && googleGroup.IdentityProviderID != "" {
					*managedCFFields[i] = append(*managedCFFields[i], zero_trust.AccessRule{
						GSuite: zero_trust.GSuiteGroupRuleGSuite{
							Email:              googleGroup.Email,
							IdentityProviderID: googleGroup.IdentityProviderID,
						},
					})
				}
			}

			for _, oktaGroups := range field.OktaGroups {
				*managedCFFields[i] = append(*managedCFFields[i], zero_trust.AccessRule{
					Okta: zero_trust.OktaGroupRuleOkta{
						IdentityProviderID: oktaGroups.IdentityProviderID,
						Name:               oktaGroups.Name,
					},
				})
			}

			for _, oidcClaim := range field.OIDCClaims {
				*managedCFFields[i] = append(*managedCFFields[i], zero_trust.AccessRule{
					SAML: zero_trust.SAMLGroupRuleSAML{
						IdentityProviderID: oidcClaim.IdentityProviderID,
						AttributeName:      oidcClaim.Name,
						AttributeValue:     oidcClaim.Value,
					},
				})
			}
			for _, ghOrgs := range field.GithubOrganizations {
				*managedCFFields[i] = append(*managedCFFields[i], zero_trust.AccessRule{
					GitHubOrganization: zero_trust.GitHubOrganizationRuleGitHubOrganization{
						IdentityProviderID: ghOrgs.IdentityProviderID,
						Name:               ghOrgs.Name,
						Team:               ghOrgs.Team,
					},
				})
			}
		}
	}
}

// +kubebuilder:object:root=true

// CloudflareAccessGroupList contains a list of CloudflareAccessGroup.
type CloudflareAccessGroupList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CloudflareAccessGroup `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CloudflareAccessGroup{}, &CloudflareAccessGroupList{})
}
