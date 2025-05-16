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
	cloudflare "github.com/cloudflare/cloudflare-go/v4"
	"github.com/cloudflare/cloudflare-go/v4/zero_trust"
)

// Requires both access rules and CF UUIDs that have been referenced by them
//
//nolint:gocognit,cyclop
func (rules *CloudFlareAccessRules) ToAccessRuleParams(resolvedCfIds ResolvedCloudflareIDs) []zero_trust.AccessRuleUnionParam {
	out := []zero_trust.AccessRuleUnionParam{}

	//
	//
	//

	for _, sTokenId := range resolvedCfIds.ServiceTokenRefCfIds {
		out = append(out, zero_trust.ServiceTokenRuleParam{
			ServiceToken: cloudflare.F(zero_trust.ServiceTokenRuleServiceTokenParam{
				TokenID: cloudflare.F(sTokenId),
			}),
		})
	}

	// TODO
	for _, groupId := range rules.AccessGroupRefs {
		out = append(out, zero_trust.GroupRuleParam{
			Group: cloudflare.F(zero_trust.GroupRuleGroupParam{
				ID: cloudflare.F(groupId),
			}),
		})
	}

	//
	//
	//

	for _, email := range rules.Emails {
		out = append(out, zero_trust.EmailRuleParam{
			Email: cloudflare.F(zero_trust.EmailRuleEmailParam{
				Email: cloudflare.F(email),
			}),
		})
	}
	for _, domain := range rules.EmailDomains {
		out = append(out, zero_trust.DomainRuleParam{
			EmailDomain: cloudflare.F(zero_trust.DomainRuleEmailDomainParam{
				Domain: cloudflare.F(domain),
			}),
		})
	}
	for _, ip := range rules.IPRanges {
		out = append(out, zero_trust.IPRuleParam{
			IP: cloudflare.F(zero_trust.IPRuleIPParam{
				IP: cloudflare.F(ip),
			}),
		})
	}

	if rules.AnyAccessServiceToken {
		out = append(out, zero_trust.AnyValidServiceTokenRuleParam{
			AnyValidServiceToken: cloudflare.F(zero_trust.AnyValidServiceTokenRuleAnyValidServiceTokenParam{}),
		})
	}
	if rules.Everyone {
		out = append(out, zero_trust.EveryoneRuleParam{
			Everyone: cloudflare.F(zero_trust.EveryoneRuleEveryoneParam{}),
		})
	}
	if rules.ValidCertificate {
		out = append(out, zero_trust.CertificateRuleParam{
			Certificate: cloudflare.F(zero_trust.CertificateRuleCertificateParam{}),
		})
	}
	for _, countryCode := range rules.Countries {
		out = append(out, zero_trust.CountryRuleParam{
			Geo: cloudflare.F(zero_trust.CountryRuleGeoParam{
				CountryCode: cloudflare.F(countryCode),
			}),
		})
	}

	for _, commonName := range rules.CommonNames {
		out = append(out, zero_trust.AccessRuleAccessCommonNameRuleParam{
			CommonName: cloudflare.F(zero_trust.AccessRuleAccessCommonNameRuleCommonNameParam{
				CommonName: cloudflare.F(commonName),
			}),
		})
	}

	for _, loginMethod := range rules.LoginMethods {
		out = append(out, zero_trust.AccessRuleAccessLoginMethodRuleParam{
			LoginMethod: cloudflare.F(zero_trust.AccessRuleAccessLoginMethodRuleLoginMethodParam{
				ID: cloudflare.F(loginMethod),
			}),
		})
	}

	for _, googleGroup := range rules.GoogleGroups {
		if googleGroup.Email != "" && googleGroup.IdentityProviderID != "" {
			out = append(out, zero_trust.GSuiteGroupRuleParam{
				GSuite: cloudflare.F(zero_trust.GSuiteGroupRuleGSuiteParam{
					Email:              cloudflare.F(googleGroup.Email),
					IdentityProviderID: cloudflare.F(googleGroup.IdentityProviderID),
				}),
			})
		}
	}

	for _, oktaGroup := range rules.OktaGroups {
		out = append(out, zero_trust.OktaGroupRuleParam{
			Okta: cloudflare.F(zero_trust.OktaGroupRuleOktaParam{
				IdentityProviderID: cloudflare.F(oktaGroup.IdentityProviderID),
				Name:               cloudflare.F(oktaGroup.Name),
			}),
		})
	}

	for _, samlGroup := range rules.SAMLGroups {
		out = append(out, zero_trust.SAMLGroupRuleParam{
			SAML: cloudflare.F(zero_trust.SAMLGroupRuleSAMLParam{
				IdentityProviderID: cloudflare.F(samlGroup.IdentityProviderID),
				AttributeName:      cloudflare.F(samlGroup.Name),
				AttributeValue:     cloudflare.F(samlGroup.Value),
			}),
		})
	}
	for _, ghOrg := range rules.GithubOrganizations {
		out = append(out, zero_trust.GitHubOrganizationRuleParam{
			GitHubOrganization: cloudflare.F(zero_trust.GitHubOrganizationRuleGitHubOrganizationParam{
				IdentityProviderID: cloudflare.F(ghOrg.IdentityProviderID),
				Name:               cloudflare.F(ghOrg.Name),
				Team:               cloudflare.F(ghOrg.Team),
			}),
		})
	}

	return out
}

//nolint:gocognit,cyclop
func (rules *CloudFlareAccessRules) ToAccessRules(resolvedCfIds ResolvedCloudflareIDs) []zero_trust.AccessRule {
	out := []zero_trust.AccessRule{}

	//
	//
	//

	for _, group := range resolvedCfIds.AccessGroupRefCfIds {
		out = append(out, zero_trust.AccessRule{
			Group: zero_trust.GroupRuleGroup{
				ID: group,
			},
		})
	}

	for _, token := range resolvedCfIds.ServiceTokenRefCfIds {
		out = append(out, zero_trust.AccessRule{
			ServiceToken: zero_trust.ServiceTokenRuleServiceToken{
				TokenID: token,
			},
		})
	}

	//
	//
	//

	for _, email := range rules.Emails {
		out = append(out, zero_trust.AccessRule{
			Email: zero_trust.EmailRuleEmail{
				Email: email,
			},
		})
	}

	for _, domain := range rules.EmailDomains {
		out = append(out, zero_trust.AccessRule{
			EmailDomain: zero_trust.DomainRuleEmailDomain{
				Domain: domain,
			},
		})
	}

	for _, ip := range rules.IPRanges {
		out = append(out, zero_trust.AccessRule{
			IP: zero_trust.IPRuleIP{
				IP: ip,
			},
		})
	}

	if rules.AnyAccessServiceToken {
		out = append(out, zero_trust.AccessRule{
			AnyValidServiceToken: zero_trust.AnyValidServiceTokenRuleAnyValidServiceToken{},
		})
	}
	if rules.Everyone {
		out = append(out, zero_trust.AccessRule{
			Everyone: zero_trust.EveryoneRuleEveryone{},
		})
	}
	if rules.ValidCertificate {
		out = append(out, zero_trust.AccessRule{
			Certificate: zero_trust.CertificateRuleCertificate{},
		})
	}

	for _, countryCode := range rules.Countries {
		out = append(out, zero_trust.AccessRule{
			Geo: zero_trust.CountryRuleGeo{
				CountryCode: countryCode,
			},
		})
	}

	for _, commonName := range rules.CommonNames {
		out = append(out, zero_trust.AccessRule{
			CommonName: zero_trust.AccessRuleAccessCommonNameRuleCommonName{
				CommonName: commonName,
			},
		})
	}

	for _, loginMethod := range rules.LoginMethods {
		out = append(out, zero_trust.AccessRule{
			LoginMethod: zero_trust.AccessRuleAccessLoginMethodRuleLoginMethod{
				ID: loginMethod,
			},
		})
	}

	for _, googleGroup := range rules.GoogleGroups {
		if googleGroup.Email != "" && googleGroup.IdentityProviderID != "" {
			out = append(out, zero_trust.AccessRule{
				GSuite: zero_trust.GSuiteGroupRuleGSuite{
					Email:              googleGroup.Email,
					IdentityProviderID: googleGroup.IdentityProviderID,
				},
			})
		}
	}

	for _, oktaGroup := range rules.OktaGroups {
		out = append(out, zero_trust.AccessRule{
			Okta: zero_trust.OktaGroupRuleOkta{
				IdentityProviderID: oktaGroup.IdentityProviderID,
				Name:               oktaGroup.Name,
			},
		})
	}

	for _, samlGroup := range rules.SAMLGroups {
		out = append(out, zero_trust.AccessRule{
			SAML: zero_trust.SAMLGroupRuleSAML{
				IdentityProviderID: samlGroup.IdentityProviderID,
				AttributeName:      samlGroup.Name,
				AttributeValue:     samlGroup.Value,
			},
		})
	}
	for _, ghOrg := range rules.GithubOrganizations {
		out = append(out, zero_trust.AccessRule{
			GitHubOrganization: zero_trust.GitHubOrganizationRuleGitHubOrganization{
				IdentityProviderID: ghOrg.IdentityProviderID,
				Name:               ghOrg.Name,
				Team:               ghOrg.Team,
			},
		})
	}

	return out
}
