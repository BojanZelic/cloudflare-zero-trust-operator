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

import "k8s.io/apimachinery/pkg/types"

// CloudFlareAccessRules defines the rules used in CloudflareAccessGroup / CloudflareAccessReusablePolicy
type CloudFlareAccessRules struct {
	// Matches specific email adresses
	//
	// +optional
	Emails []string `json:"emails,omitempty"`

	// Matches specific email domains
	//
	// +optional
	EmailDomains []string `json:"emailDomains,omitempty"`

	// Matches IP CIDR blocks (https://www.ipaddressguide.com/cidr)
	//
	// +optional
	IPRanges []string `json:"ipRanges,omitempty"`

	// Matches Country IDs (https://en.wikipedia.org/wiki/ISO_3166-1_alpha-2)
	//
	// +optional
	Countries []string `json:"countries,omitempty"`

	// Matches on Identity Provider UUIDs (https://developers.cloudflare.com/api/resources/zero_trust/subresources/identity_providers/methods/get/)
	//
	// +optional
	LoginMethods []string `json:"loginMethods,omitempty"`

	// Matches Certificates CNs
	//
	// +optional
	CommonNames []string `json:"commonNames,omitempty"`

	// Allow Everyone; would always be a match
	//
	// +optional
	Everyone bool `json:"everyone,omitzero"`

	// Would be a match if using any valid certificate
	//
	// +optional
	ValidCertificate bool `json:"validCertificate,omitzero"`

	// Would be a match if using any valid service token
	//
	// +optional
	AnyAccessServiceToken bool `json:"anyAccessServiceToken,omitzero"`

	// Would match access groups refs by {name} or {namespace/name} of [CloudflareServiceToken]
	//
	// +optional
	ServiceTokenRefs []string `json:"serviceTokenRefs,omitempty"`

	// Would match access groups refs by {name} or {namespace/name} of [CloudflareAccessGroup]
	//
	// +optional
	AccessGroupRefs []string `json:"accessGroupRefs,omitempty"`

	// Matches Google Groups
	//
	// +optional
	GoogleGroups []GoogleGroup `json:"googleGroups,omitempty"`

	// Matches Okta Groups
	//
	// +optional
	OktaGroups []OktaGroup `json:"oktaGroups,omitempty"`

	// Matches SAML Groups
	//
	// +optional
	SAMLGroups []SAMLGroup `json:"samlGroups,omitempty"`

	// Matches Github Organizations
	//
	// +optional
	GithubOrganizations []GithubOrganization `json:"githubOrganizations,omitempty"`
}

//
//
//

type RulerResolvedCloudflareIDs struct {
	// +required
	Include ResolvedCloudflareIDs `json:"include,omitzero"`
	// +optional
	Require ResolvedCloudflareIDs `json:"require,omitzero"`
	// +optional
	Exclude ResolvedCloudflareIDs `json:"exclude,omitzero"`
}

type ResolvedCloudflareIDs struct {
	// +optional
	AccessGroupRefCfIds []string `json:"accessGroupRefCfIds,omitempty"`
	// +optional
	ServiceTokenRefCfIds []string `json:"serviceTokenRefCfIds,omitempty"`
}

//
//
//

func (rules *CloudFlareAccessRules) GetNamespacedServiceTokenRefs(contextNamespace string) ([]types.NamespacedName, error) {
	return parseNamespacedNames(rules.ServiceTokenRefs, contextNamespace)
}

func (rules *CloudFlareAccessRules) GetNamespacedGroupRefs(contextNamespace string) ([]types.NamespacedName, error) {
	return parseNamespacedNames(rules.AccessGroupRefs, contextNamespace)
}
