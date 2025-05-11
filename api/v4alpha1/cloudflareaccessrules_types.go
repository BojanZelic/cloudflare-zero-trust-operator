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
	Emails []string `json:"emails,omitempty"`

	// Matches specific email domains
	EmailDomains []string `json:"emailDomains,omitempty"`

	// Matches IP CIDR blocks (https://www.ipaddressguide.com/cidr)
	IPRanges []string `json:"ipRanges,omitempty"`

	// Matches Country IDs (https://en.wikipedia.org/wiki/ISO_3166-1_alpha-2)
	Countries []string `json:"countries,omitempty"`

	// Matches on Identity Provider UUIDs (https://developers.cloudflare.com/api/resources/zero_trust/subresources/identity_providers/methods/get/)
	LoginMethods []string `json:"loginMethods,omitempty"`

	// Matches Certificates CNs
	CommonNames []string `json:"commonNames,omitempty"`

	// Allow Everyone; would always be a match
	Everyone *bool `json:"everyone,omitempty"`

	// Would be a match if using any valid certificate
	ValidCertificate *bool `json:"validCertificate,omitempty"`

	// Would be a match if using any valid service token
	AnyAccessServiceToken *bool `json:"anyAccessServiceToken,omitempty"`

	// Would match access groups refs by {name} or {namespace/name} of [CloudflareServiceToken]
	ServiceTokenRefs []string `json:"serviceTokenRefs,omitempty"`

	// Would match access groups refs by {name} or {namespace/name} of [CloudflareAccessGroup]
	GroupRefs []string `json:"accessGroupRefs,omitempty"`

	// Matches Google Groups
	GoogleGroups []GoogleGroup `json:"googleGroups,omitempty"`

	// Matches Okta Groups
	OktaGroups []OktaGroup `json:"oktaGroups,omitempty"`

	// Matches SAML Groups
	SAMLGroups []SAMLGroup `json:"samlGroups,omitempty"`

	// Matches Github Organizations
	GithubOrganizations []GithubOrganization `json:"githubOrganizations,omitempty"`
}

//
//
//

type RulerResolvedCloudflareIDs struct {
	// +optional
	Include ResolvedCloudflareIDs `json:"include"`
	// +optional
	Require ResolvedCloudflareIDs `json:"require"`
	// +optional
	Exclude ResolvedCloudflareIDs `json:"exclude"`
}

type ResolvedCloudflareIDs struct {
	// +optional
	GroupRefCfIds []string `json:"groupRefCfIds,omitempty"`
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
	return parseNamespacedNames(rules.GroupRefs, contextNamespace)
}
