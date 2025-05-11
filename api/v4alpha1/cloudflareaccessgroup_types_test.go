package v4alpha1_test

import (
	"encoding/json"
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
			accessGroup.Spec.Include = v4alpha1.CloudFlareAccessRules{{
				Emails: emails,
			}}
			accessGroup.Spec.Require = v4alpha1.CloudFlareAccessRules{{
				Emails: emails,
			}}
			accessGroup.Spec.Exclude = v4alpha1.CloudFlareAccessRules{{
				Emails: emails,
			}}

			for i := range emails { //nolint:varnamelen
				include, ok := json.Marshal(accessGroup.Spec.Include[i])
				exclude, ok := json.Marshal(accessGroup.Spec.Exclude[i])
				require, ok := json.Marshal(accessGroup.Spec.Require[i])

				includeAR, ok := json.Marshal(accessGroup.Spec.Include[i])
				excludeAR, ok := json.Marshal(accessGroup.Spec.Exclude[i])
				requireAR, ok := json.Marshal(accessGroup.Spec.Require[i])

				Expect(include).To(Equal(
					zero_trust.AccessRule{
						Email: zero_trust.EmailRuleEmail{
							Email: emails[i],
						},
					},
				))
				Expect(accessGroup.ToCloudflare().Require[i]).To(Equal(
					zero_trust.AccessRule{
						Email: zero_trust.EmailRuleEmail{
							Email: emails[i],
						},
					},
				))
				Expect(accessGroup.ToCloudflare().Exclude[i]).To(Equal(
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
			accessGroup.Spec.Include = []v4alpha1.CloudFlareAccessRules{{
				EmailDomains: domains,
			}}
			for i := range domains {
				Expect(accessGroup.ToCloudflare().Include[i]).To(Equal(
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
			accessGroup.Spec.Include = []v4alpha1.CloudFlareAccessRules{{
				GoogleGroups: googleGroups,
			}}
			for i, group := range googleGroups {
				Expect(accessGroup.ToCloudflare().Include[i]).To(Equal(
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
			accessGroup.Spec.Include = []v4alpha1.CloudFlareAccessRules{{
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
			for i, group := range accessGroup.Spec.Include[0].OktaGroups {
				Expect(accessGroup.ToCloudflare().Include[i]).To(Equal(
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
			accessGroup.Spec.Include = []v4alpha1.CloudFlareAccessRules{{
				SAMLGroups: []v4alpha1.SAMLGroup{
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
				},
			}}
			for i, group := range accessGroup.Spec.Include[0].SAMLGroups {
				Expect(accessGroup.ToCloudflare().Include[i]).To(Equal(
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
			accessGroup.Spec.Include = []v4alpha1.CloudFlareAccessRules{{
				IPRanges: ips},
			}
			for i := range ips {
				Expect(accessGroup.ToCloudflare().Include[i]).To(Equal(
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
			accessGroup.Spec.Include = []v4alpha1.CloudFlareAccessRules{{
				ServiceTokens: ids,
			}}
			for i, id := range ids {
				Expect(accessGroup.ToCloudflare().Include[i]).To(Equal(
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
			accessGroup.Spec.Include = []v4alpha1.CloudFlareAccessRules{{
				AnyAccessServiceToken: &validServiceToken,
			}}
			Expect(accessGroup.ToCloudflare().Include[0]).To(Equal(
				zero_trust.AccessRule{
					AnyValidServiceToken: zero_trust.AnyValidServiceTokenRuleAnyValidServiceToken{},
				},
			))
		})

		It("can export validCertificate to the cloudflare object", func() {
			validCert := true
			accessGroup.Spec.Include = []v4alpha1.CloudFlareAccessRules{{
				ValidCertificate: &validCert,
			}}
			Expect(accessGroup.ToCloudflare().Include[0]).To(Equal(
				zero_trust.AccessRule{
					Certificate: zero_trust.CertificateRuleCertificate{},
				},
			))
		})

		It("can export everyone to the cloudflare object", func() {
			everyone := true
			accessGroup.Spec.Include = []v4alpha1.CloudFlareAccessRules{{
				Everyone: &everyone,
			}}
			Expect(accessGroup.ToCloudflare().Include[0]).To(Equal(
				zero_trust.AccessRule{
					Everyone: zero_trust.EveryoneRuleEveryone{},
				},
			))
		})

		It("can export Country to the cloudflare object", func() {
			accessGroup.Spec.Include = []v4alpha1.CloudFlareAccessRules{{
				Countries: []string{"US", "UK"},
			}}
			for i, country := range accessGroup.Spec.Include[0].Countries {
				Expect(accessGroup.ToCloudflare().Include[i]).To(Equal(
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
			accessGroup.Spec.Include = []v4alpha1.CloudFlareAccessRules{{
				AccessGroups: ids,
			}}
			for i, id := range ids {
				Expect(accessGroup.ToCloudflare().Include[i]).To(Equal(
					zero_trust.AccessRule{
						Group: zero_trust.GroupRuleGroup{
							ID: id.Value,
						},
					},
				))
			}
		})

		It("can export loginMethods to the cloudflare object", func() {
			accessGroup.Spec.Include = []v4alpha1.CloudFlareAccessRules{{
				LoginMethods: []string{"00000000-1234-5678-1234-123456789012"},
			}}

			for i, id := range accessGroup.Spec.Include[0].LoginMethods {
				Expect(accessGroup.ToCloudflare().Include[i]).To(Equal(
					zero_trust.AccessRule{
						LoginMethod: zero_trust.AccessRuleAccessLoginMethodRuleLoginMethod{
							ID: id,
						},
					},
				))
			}
		})

		It("can export github organizations to the cloudflare object", func() {
			accessGroup.Spec.Include = []v4alpha1.CloudFlareAccessRules{{
				GithubOrganizations: []v4alpha1.GithubOrganization{{
					Name:               "test",
					IdentityProviderID: "zelic-io",
					Team:               "dev",
				}},
			}}

			for i, org := range accessGroup.Spec.Include[0].GithubOrganizations {
				Expect(accessGroup.ToCloudflare().Include[i]).To(Equal(
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
