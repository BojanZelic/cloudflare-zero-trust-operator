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
	var accessRule *v4alpha1.CloudflareAccessGroup

	When("the instance is created", func() {
		BeforeEach(func() {
			accessRule = &v4alpha1.CloudflareAccessGroup{}
		})

		It("can export Included emails to the cloudflare object", func() {
			emails := []string{"test@email.com", "test2@email.com"}
			accessRule.Spec.Include = []v4alpha1.CloudFlareAccessRule{{
				Emails: emails,
			}}
			accessRule.Spec.Require = []v4alpha1.CloudFlareAccessRule{{
				Emails: emails,
			}}
			accessRule.Spec.Exclude = []v4alpha1.CloudFlareAccessRule{{
				Emails: emails,
			}}

			for i := range emails { //nolint:varnamelen
				Expect(accessRule.ToCloudflare().Include[i]).To(Equal(
					zero_trust.AccessRule{
						Email: zero_trust.EmailRuleEmail{
							Email: emails[i],
						},
					},
				))
				Expect(accessRule.ToCloudflare().Require[i]).To(Equal(
					zero_trust.AccessRule{
						Email: zero_trust.EmailRuleEmail{
							Email: emails[i],
						},
					},
				))
				Expect(accessRule.ToCloudflare().Exclude[i]).To(Equal(
					zero_trust.AccessRule{
						Email: zero_trust.EmailRuleEmail{
							Email: emails[i],
						},
					},
				))
			}
		})

		It("can export emaildomains to the cloudflare object", func() {
			domains := []string{"email.com", "email2.com"}
			accessRule.Spec.Include = []v4alpha1.CloudFlareAccessRule{{
				EmailDomains: domains,
			}}
			for i := range domains {
				Expect(accessRule.ToCloudflare().Include[i]).To(Equal(
					zero_trust.AccessRule{
						EmailDomain: zero_trust.DomainRuleEmailDomain{
							Domain: domains[i],
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
			accessRule.Spec.Include = []v4alpha1.CloudFlareAccessRule{{
				GoogleGroups: googleGroups,
			}}
			for i, group := range googleGroups {
				Expect(accessRule.ToCloudflare().Include[i]).To(Equal(
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
			accessRule.Spec.Include = []v4alpha1.CloudFlareAccessRule{{
				OktaGroups: []v4alpha1.OktaGroup{
					{
						Name:               "myOktaGroup1",
						IdentityProviderID: "00000000-0000-0000-0000-00000000000000",
					},
					{
						Name:               "myOktaGroup2",
						IdentityProviderID: "11111111-1111-1111-1111-111111111111",
					},
				},
			}}
			for i, group := range accessRule.Spec.Include[0].OktaGroups {
				Expect(accessRule.ToCloudflare().Include[i]).To(Equal(
					zero_trust.AccessRule{
						Okta: zero_trust.OktaGroupRuleOkta{
							Name:               group.Name,
							IdentityProviderID: group.IdentityProviderID,
						},
					},
				))
			}
		})

		It("can export oidcClaims to the cloudflare object", func() {
			accessRule.Spec.Include = []v4alpha1.CloudFlareAccessRule{{
				OIDCClaims: []v4alpha1.OIDCClaim{
					{
						Name:               "myOidcClaimName1",
						Value:              "myOidcClaimValue1",
						IdentityProviderID: "00000000-0000-0000-0000-00000000000000",
					},
					{
						Name:               "myOidcClaimName2",
						Value:              "myOidcClaimValue2",
						IdentityProviderID: "11111111-1111-1111-1111-111111111111",
					},
				},
			}}
			for i, group := range accessRule.Spec.Include[0].OIDCClaims {
				Expect(accessRule.ToCloudflare().Include[i]).To(Equal(
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
			accessRule.Spec.Include = []v4alpha1.CloudFlareAccessRule{{
				IPRanges: ips},
			}
			for i := range ips {
				Expect(accessRule.ToCloudflare().Include[i]).To(Equal(
					zero_trust.AccessRule{
						IP: zero_trust.IPRuleIP{
							IP: ips[i],
						},
					},
				))
			}
		})

		It("can export serviceTokens to the cloudflare object", func() {
			ids := []v4alpha1.ServiceToken{{Value: "some_service_token"}, {Value: "some_other_service_token"}}
			accessRule.Spec.Include = []v4alpha1.CloudFlareAccessRule{{
				ServiceTokens: ids,
			}}
			for i, id := range ids {
				Expect(accessRule.ToCloudflare().Include[i]).To(Equal(
					zero_trust.AccessRule{
						ServiceToken: zero_trust.ServiceTokenRuleServiceToken{
							TokenID: id.Value,
						},
					},
				))
			}
		})

		It("can export any serviceTokens to the cloudflare object", func() {
			validServiceToken := true
			accessRule.Spec.Include = []v4alpha1.CloudFlareAccessRule{{
				AnyAccessServiceToken: &validServiceToken,
			}}
			Expect(accessRule.ToCloudflare().Include[0]).To(Equal(
				zero_trust.AccessRule{
					AnyValidServiceToken: zero_trust.AnyValidServiceTokenRuleAnyValidServiceToken{},
				},
			))
		})

		It("can export validCertificate to the cloudflare object", func() {
			validCert := true
			accessRule.Spec.Include = []v4alpha1.CloudFlareAccessRule{{
				ValidCertificate: &validCert,
			}}
			Expect(accessRule.ToCloudflare().Include[0]).To(Equal(
				zero_trust.AccessRule{
					Certificate: zero_trust.CertificateRuleCertificate{},
				},
			))
		})

		It("can export everyone to the cloudflare object", func() {
			everyone := true
			accessRule.Spec.Include = []v4alpha1.CloudFlareAccessRule{{
				Everyone: &everyone,
			}}
			Expect(accessRule.ToCloudflare().Include[0]).To(Equal(
				zero_trust.AccessRule{
					Everyone: zero_trust.EveryoneRuleEveryone{},
				},
			))
		})

		It("can export Country to the cloudflare object", func() {
			accessRule.Spec.Include = []v4alpha1.CloudFlareAccessRule{{
				Countries: []string{"US", "UK"},
			}}
			for i, country := range accessRule.Spec.Include[0].Countries {
				Expect(accessRule.ToCloudflare().Include[i]).To(Equal(
					zero_trust.AccessRule{
						Geo: zero_trust.CountryRuleGeo{
							CountryCode: country,
						},
					},
				))
			}
		})

		It("can export accessGroups to the cloudflare object", func() {
			ids := []v4alpha1.AccessGroup{{Value: "first_access_group_id"}, {Value: "second_access_group_id"}}
			accessRule.Spec.Include = []v4alpha1.CloudFlareAccessRule{{
				AccessGroups: ids,
			}}
			for i, id := range ids {
				Expect(accessRule.ToCloudflare().Include[i]).To(Equal(
					zero_trust.AccessRule{
						Group: zero_trust.GroupRuleGroup{
							ID: id.Value,
						},
					},
				))
			}
		})

		It("can export loginMethods to the cloudflare object", func() {
			accessRule.Spec.Include = []v4alpha1.CloudFlareAccessRule{{
				LoginMethods: []string{"00000000-1234-5678-1234-123456789012"},
			}}

			for i, id := range accessRule.Spec.Include[0].LoginMethods {
				Expect(accessRule.ToCloudflare().Include[i]).To(Equal(
					zero_trust.AccessRule{
						LoginMethod: zero_trust.AccessRuleAccessLoginMethodRuleLoginMethod{
							ID: id,
						},
					},
				))
			}
		})

		It("can export github organizations to the cloudflare object", func() {
			accessRule.Spec.Include = []v4alpha1.CloudFlareAccessRule{{
				GithubOrganizations: []v4alpha1.GithubOrganization{{
					Name:               "test",
					IdentityProviderID: "zelic-io",
					Team:               "dev",
				}},
			}}

			for i, org := range accessRule.Spec.Include[0].GithubOrganizations {
				Expect(accessRule.ToCloudflare().Include[i]).To(Equal(
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
