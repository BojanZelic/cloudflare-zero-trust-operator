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

// CloudFlareAccessRule defines the rules used in CloudflareAccessGroup
type CloudFlareAccessRule struct {
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
