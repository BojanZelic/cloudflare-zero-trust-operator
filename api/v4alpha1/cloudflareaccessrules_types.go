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

// CloudFlareAccessRules defines the rules used in CloudflareAccessGroup / CloudflareAccessReusablePolicy
type CloudFlareAccessRules struct {
	// Matches specific emails
	Emails []string `json:"emails,omitempty"`

	// Matches a specific email domains
	EmailDomains []string `json:"emailDomains,omitempty"`

	// Matches IP CIDR blocks
	IPRanges []string `json:"ipRanges,omitempty"`

	// Matches Country IDs
	Countries []string `json:"countries,omitempty"`

	// Matches ID of login methods
	LoginMethods []string `json:"loginMethods,omitempty"`

	// Matches Certificate CNs
	CommonNames []string `json:"commonNames,omitempty"`

	// Allow Everyone
	Everyone *bool `json:"everyone,omitempty"`

	// Matches Any valid certificate
	ValidCertificate *bool `json:"validCertificate,omitempty"`

	// Matches any valid service token
	AnyAccessServiceToken *bool `json:"anyAccessServiceToken,omitempty"`

	// Matches service tokens
	ServiceTokens []ServiceToken `json:"serviceTokens,omitempty"`

	// Would match other access groups
	AccessGroups []AccessGroup `json:"accessGroups,omitempty"`

	// Matches Google Groups
	GoogleGroups []GoogleGroup `json:"googleGroups,omitempty"`

	// Matches Okta Groups
	OktaGroups []OktaGroup `json:"oktaGroups,omitempty"`

	// Matches OIDC Claims
	OIDCClaims []OIDCClaim `json:"oidcClaims,omitempty"`

	// Matches Github Organizations
	GithubOrganizations []GithubOrganization `json:"githubOrganizations,omitempty"`
}
