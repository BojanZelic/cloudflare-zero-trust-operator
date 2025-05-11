package v4alpha1_test

import (
	"testing"

	"github.com/bojanzelic/cloudflare-zero-trust-operator/api/v4alpha1"
	"github.com/cloudflare/cloudflare-go/v4/zero_trust"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	// +kubebuilder:scaffold:imports
)

func TestBooks(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "CloudflareAccessGroup Suite")
}

var _ = Describe("Creating a CloudflareAccessGroup", Label("CloudflareAccessGroup"), func() {
	var accessGroup *v4alpha1.CloudflareAccessGroup

	When("the instance is created", func() {
		BeforeEach(func() {
			accessGroup = &v4alpha1.CloudflareAccessGroup{}
		})

		It("can export Included emails to the cloudflare object", func() {
			emails := []string{"test@email.com", "test2@email.com"}
			accessGroup.Spec.Include = v4alpha1.CloudFlareAccessRules{
				Emails: emails,
			}
			accessGroup.Spec.Require = v4alpha1.CloudFlareAccessRules{
				Emails: emails,
			}
			accessGroup.Spec.Exclude = v4alpha1.CloudFlareAccessRules{
				Emails: emails,
			}

			include := accessGroup.Spec.Include.ToAccessRules(accessGroup.Status.ResolvedIdpsFromRefs.Include)
			exclude := accessGroup.Spec.Exclude.ToAccessRules(accessGroup.Status.ResolvedIdpsFromRefs.Exclude)
			require := accessGroup.Spec.Require.ToAccessRules(accessGroup.Status.ResolvedIdpsFromRefs.Require)

			for i, email := range emails { //nolint:varnamelen
				Expect(include[i]).To(Equal(
					zero_trust.AccessRule{
						Email: zero_trust.EmailRuleEmail{
							Email: email,
						},
					},
				))
				Expect(require[i]).To(Equal(
					zero_trust.AccessRule{
						Email: zero_trust.EmailRuleEmail{
							Email: email,
						},
					},
				))
				Expect(exclude[i]).To(Equal(
					zero_trust.AccessRule{
						Email: zero_trust.EmailRuleEmail{
							Email: email,
						},
					},
				))
			}
		})

		It("can export emaildomains to the cloudflare object", func() {
			domains := []string{"email.com", "email2.com"}
			accessGroup.Spec.Include = v4alpha1.CloudFlareAccessRules{
				EmailDomains: domains,
			}

			include := accessGroup.Spec.Include.ToAccessRules(accessGroup.Status.ResolvedIdpsFromRefs.Include)

			for i, domain := range domains {
				Expect(include[i]).To(Equal(
					zero_trust.AccessRule{
						EmailDomain: zero_trust.DomainRuleEmailDomain{
							Domain: domain,
						},
					},
				))
			}
		})

		It("can export googleGroups to the cloudflare object", func() {
			googleGroups := []v4alpha1.GoogleGroup{
				{
					Email:              "test@email.com",
					IdentityProviderID: "00000000-0000-0000-0000-00000000000000",
				},
				{
					Email:              "test2@email.com",
					IdentityProviderID: "11111111-1111-1111-1111-111111111111",
				},
			}
			accessGroup.Spec.Include = v4alpha1.CloudFlareAccessRules{
				GoogleGroups: googleGroups,
			}

			include := accessGroup.Spec.Include.ToAccessRules(accessGroup.Status.ResolvedIdpsFromRefs.Include)

			for i, group := range googleGroups {
				Expect(include[i]).To(Equal(
					zero_trust.AccessRule{
						GSuite: zero_trust.GSuiteGroupRuleGSuite{
							Email:              group.Email,
							IdentityProviderID: group.IdentityProviderID,
						},
					},
				))
			}
		})

		It("can export oktaGroups to the cloudflare object", func() {
			oktaGroups := []v4alpha1.OktaGroup{
				{
					Name:               "myOktaGroup1",
					IdentityProviderID: "00000000-0000-0000-0000-00000000000000",
				},
				{
					Name:               "myOktaGroup2",
					IdentityProviderID: "11111111-1111-1111-1111-111111111111",
				},
			}

			accessGroup.Spec.Include = v4alpha1.CloudFlareAccessRules{
				OktaGroups: oktaGroups,
			}

			include := accessGroup.Spec.Include.ToAccessRules(accessGroup.Status.ResolvedIdpsFromRefs.Include)

			for i, group := range oktaGroups {
				Expect(include[i]).To(Equal(
					zero_trust.AccessRule{
						Okta: zero_trust.OktaGroupRuleOkta{
							Name:               group.Name,
							IdentityProviderID: group.IdentityProviderID,
						},
					},
				))
			}
		})

		It("can export samlGroups to the cloudflare object", func() {
			samlGroups := []v4alpha1.SAMLGroup{
				{
					Name:               "mySamlGroupName1",
					Value:              "mySamlGroupValue1",
					IdentityProviderID: "00000000-0000-0000-0000-00000000000000",
				},
				{
					Name:               "mySamlGroupName2",
					Value:              "mySamlGroupValue2",
					IdentityProviderID: "11111111-1111-1111-1111-111111111111",
				},
			}

			accessGroup.Spec.Include = v4alpha1.CloudFlareAccessRules{
				SAMLGroups: samlGroups,
			}

			include := accessGroup.Spec.Include.ToAccessRules(accessGroup.Status.ResolvedIdpsFromRefs.Include)

			for i, group := range samlGroups {
				Expect(include[i]).To(Equal(
					zero_trust.AccessRule{
						SAML: zero_trust.SAMLGroupRuleSAML{
							AttributeName:      group.Name,
							AttributeValue:     group.Value,
							IdentityProviderID: group.IdentityProviderID,
						},
					},
				))
			}
		})

		It("can export ipRanges to the cloudflare object", func() {
			ips := []string{"1.1.1.1/32", "8.8.8.8/32"}
			accessGroup.Spec.Include = v4alpha1.CloudFlareAccessRules{
				IPRanges: ips,
			}

			include := accessGroup.Spec.Include.ToAccessRules(accessGroup.Status.ResolvedIdpsFromRefs.Include)

			for i, ip := range ips {
				Expect(include[i]).To(Equal(
					zero_trust.AccessRule{
						IP: zero_trust.IPRuleIP{
							IP: ip,
						},
					},
				))
			}
		})

		It("can export serviceTokens to the cloudflare object", func() {
			// below is for illustration only
			refs := []string{"service-token-1", "service-token-2"}
			accessGroup.Spec.Include = v4alpha1.CloudFlareAccessRules{
				ServiceTokenRefs: refs,
			}

			// these would be resolved CloudFlare UUIDs of above underlying resources
			refIds := []string{"00001100-1234-5678-1234-123456789012", "00000000-1334-5678-1234-123456789012"}
			accessGroup.Status.ResolvedIdpsFromRefs.Include = v4alpha1.ResolvedCloudflareIDs{
				ServiceTokenRefCfIds: refIds,
			}

			include := accessGroup.Spec.Include.ToAccessRules(accessGroup.Status.ResolvedIdpsFromRefs.Include)

			for i, serviceTokenRefId := range refIds {
				Expect(include[i]).To(Equal(
					zero_trust.AccessRule{
						ServiceToken: zero_trust.ServiceTokenRuleServiceToken{
							TokenID: serviceTokenRefId,
						},
					},
				))
			}
		})

		It("can export accessGroups to the cloudflare object", func() {
			// bellow is for illustration only
			refs := []string{"access-group-1", "access-group-2"}
			accessGroup.Spec.Include = v4alpha1.CloudFlareAccessRules{
				GroupRefs: refs,
			}

			// these would be resolved CloudFlare UUIDs of above underlying resources
			refIds := []string{"000441100-1234-5678-1234-123456789012", "00004200-1334-5678-1234-123456789012"}
			accessGroup.Status.ResolvedIdpsFromRefs.Include = v4alpha1.ResolvedCloudflareIDs{
				GroupRefCfIds: refIds,
			}

			include := accessGroup.Spec.Include.ToAccessRules(accessGroup.Status.ResolvedIdpsFromRefs.Include)

			for i, groupRefId := range refIds {
				Expect(include[i]).To(Equal(
					zero_trust.AccessRule{
						Group: zero_trust.GroupRuleGroup{
							ID: groupRefId,
						},
					},
				))
			}
		})

		It("can export any serviceTokens to the cloudflare object", func() {
			validServiceToken := true
			accessGroup.Spec.Include = v4alpha1.CloudFlareAccessRules{
				AnyAccessServiceToken: &validServiceToken,
			}

			include := accessGroup.Spec.Include.ToAccessRules(accessGroup.Status.ResolvedIdpsFromRefs.Include)

			Expect(include[0]).To(Equal(
				zero_trust.AccessRule{
					AnyValidServiceToken: zero_trust.AnyValidServiceTokenRuleAnyValidServiceToken{},
				},
			))
		})

		It("can export validCertificate to the cloudflare object", func() {
			validCert := true
			accessGroup.Spec.Include = v4alpha1.CloudFlareAccessRules{
				ValidCertificate: &validCert,
			}

			include := accessGroup.Spec.Include.ToAccessRules(accessGroup.Status.ResolvedIdpsFromRefs.Include)

			Expect(include[0]).To(Equal(
				zero_trust.AccessRule{
					Certificate: zero_trust.CertificateRuleCertificate{},
				},
			))
		})

		It("can export everyone to the cloudflare object", func() {
			everyone := true
			accessGroup.Spec.Include = v4alpha1.CloudFlareAccessRules{
				Everyone: &everyone,
			}

			include := accessGroup.Spec.Include.ToAccessRules(accessGroup.Status.ResolvedIdpsFromRefs.Include)

			Expect(include[0]).To(Equal(
				zero_trust.AccessRule{
					Everyone: zero_trust.EveryoneRuleEveryone{},
				},
			))
		})

		It("can export Country to the cloudflare object", func() {
			countries := []string{"US", "UK"}
			accessGroup.Spec.Include = v4alpha1.CloudFlareAccessRules{
				Countries: countries,
			}

			include := accessGroup.Spec.Include.ToAccessRules(accessGroup.Status.ResolvedIdpsFromRefs.Include)

			for i, country := range countries {
				Expect(include[i]).To(Equal(
					zero_trust.AccessRule{
						Geo: zero_trust.CountryRuleGeo{
							CountryCode: country,
						},
					},
				))
			}
		})

		It("can export loginMethods to the cloudflare object", func() {
			loginMethodsIdps := []string{"00000000-1234-5678-1234-123456789012"}

			accessGroup.Spec.Include = v4alpha1.CloudFlareAccessRules{
				LoginMethods: loginMethodsIdps,
			}

			include := accessGroup.Spec.Include.ToAccessRules(accessGroup.Status.ResolvedIdpsFromRefs.Include)

			for i, idProvider := range loginMethodsIdps {
				Expect(include[i]).To(Equal(
					zero_trust.AccessRule{
						LoginMethod: zero_trust.AccessRuleAccessLoginMethodRuleLoginMethod{
							ID: idProvider,
						},
					},
				))
			}
		})

		It("can export github organizations to the cloudflare object", func() {
			ghOrgs := []v4alpha1.GithubOrganization{{
				Name:               "test",
				IdentityProviderID: "00000022-1234-5678-1234-123456789012",
				Team:               "dev",
			}}

			accessGroup.Spec.Include = v4alpha1.CloudFlareAccessRules{
				GithubOrganizations: ghOrgs,
			}

			include := accessGroup.Spec.Include.ToAccessRules(accessGroup.Status.ResolvedIdpsFromRefs.Include)

			for i, org := range ghOrgs {
				Expect(include[i]).To(Equal(
					zero_trust.AccessRule{
						GitHubOrganization: zero_trust.GitHubOrganizationRuleGitHubOrganization{
							IdentityProviderID: org.IdentityProviderID,
							Name:               org.Name,
							Team:               org.Team,
						},
					},
				))
			}
		})
	})
})
