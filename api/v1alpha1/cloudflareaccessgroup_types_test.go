package v1alpha1_test

import (
	"testing"

	"github.com/bojanzelic/cloudflare-zero-trust-operator/internal/cfapi"

	"github.com/bojanzelic/cloudflare-zero-trust-operator/api/v1alpha1"
	cloudflare "github.com/cloudflare/cloudflare-go"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	// +kubebuilder:scaffold:imports
)

func TestBooks(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "CloudflareAccessGroup Suite")
}

var _ = Describe("Creating a CloudflareAccessGroup", Label("CloudflareAccessGroup"), func() {
	var accessRule *v1alpha1.CloudflareAccessGroup

	When("the instance is created", func() {
		BeforeEach(func() {
			accessRule = &v1alpha1.CloudflareAccessGroup{}
		})

		It("can export Included emails to the cloudflare object", func() {
			emails := []string{"test@email.com", "test2@email.com"}
			accessRule.Spec.Include = []v1alpha1.CloudFlareAccessGroupRule{{
				Emails: emails},
			}
			accessRule.Spec.Require = []v1alpha1.CloudFlareAccessGroupRule{{
				Emails: emails},
			}
			accessRule.Spec.Exclude = []v1alpha1.CloudFlareAccessGroupRule{{
				Emails: emails},
			}
			for i := range emails {
				Expect(accessRule.ToCloudflare().Include[i]).To(Equal(cloudflare.AccessGroupEmail{
					Email: struct {
						Email string "json:\"email\""
					}{
						Email: emails[i],
					},
				}))
				Expect(accessRule.ToCloudflare().Require[i]).To(Equal(cloudflare.AccessGroupEmail{
					Email: struct {
						Email string "json:\"email\""
					}{
						Email: emails[i],
					},
				}))
				Expect(accessRule.ToCloudflare().Exclude[i]).To(Equal(cloudflare.AccessGroupEmail{
					Email: struct {
						Email string "json:\"email\""
					}{
						Email: emails[i],
					},
				}))
			}
		})

		It("can export emaildomains to the cloudflare object", func() {
			domains := []string{"email.com", "email2.com"}
			accessRule.Spec.Include = []v1alpha1.CloudFlareAccessGroupRule{{
				EmailDomains: domains},
			}
			for i := range domains {
				Expect(accessRule.ToCloudflare().Include[i]).To(Equal(cloudflare.AccessGroupEmailDomain{
					EmailDomain: struct {
						Domain string "json:\"domain\""
					}{
						Domain: domains[i],
					},
				}))
			}
		})

		It("can export googleGroups to the cloudflare object", func() {
			googleGroups := []v1alpha1.GoogleGroup{
				{
					Email:              "test@email.com",
					IdentityProviderID: "00000000-0000-0000-0000-00000000000000",
				},
				{
					Email:              "test2@email.com",
					IdentityProviderID: "11111111-1111-1111-1111-111111111111",
				},
			}
			accessRule.Spec.Include = []v1alpha1.CloudFlareAccessGroupRule{{
				GoogleGroups: googleGroups},
			}
			for i := range googleGroups {
				Expect(accessRule.ToCloudflare().Include[i]).To(Equal(cloudflare.AccessGroupGSuite{
					Gsuite: struct {
						Email              string "json:\"email\""
						IdentityProviderID string "json:\"identity_provider_id\""
					}{
						Email:              googleGroups[i].Email,
						IdentityProviderID: googleGroups[i].IdentityProviderID,
					},
				}))
			}
		})

		It("can export oktaGroups to the cloudflare object", func() {
			accessRule.Spec.Include = []v1alpha1.CloudFlareAccessGroupRule{{
				OktaGroup: []v1alpha1.OktaGroup{
					{
						Name:               "myOktaGroup1",
						IdentityProviderID: "00000000-0000-0000-0000-00000000000000",
					},
					{
						Name:               "myOktaGroup2",
						IdentityProviderID: "11111111-1111-1111-1111-111111111111",
					},
				}},
			}
			for i, group := range accessRule.Spec.Include[0].OktaGroup {
				Expect(accessRule.ToCloudflare().Include[i]).To(Equal(cloudflare.AccessGroupOkta{
					Okta: struct {
						Name               string "json:\"name\""
						IdentityProviderID string "json:\"identity_provider_id\""
					}{
						Name:               group.Name,
						IdentityProviderID: group.IdentityProviderID,
					},
				}))
			}
		})

		It("can export oidcClaims to the cloudflare object", func() {
			accessRule.Spec.Include = []v1alpha1.CloudFlareAccessGroupRule{{
				OIDCClaims: []v1alpha1.OIDCClaim{
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
				}},
			}
			for i, group := range accessRule.Spec.Include[0].OIDCClaims {
				Expect(accessRule.ToCloudflare().Include[i]).To(Equal(cfapi.AccessGroupOIDCClaim{
					OIDC: struct {
						Name               string "json:\"claim_name\""
						Value              string "json:\"claim_value\""
						IdentityProviderID string "json:\"identity_provider_id\""
					}{
						Name:               group.Name,
						Value:              group.Value,
						IdentityProviderID: group.IdentityProviderID,
					},
				}))
			}
		})

		It("can export ipRanges to the cloudflare object", func() {
			ips := []string{"1.1.1.1/32", "8.8.8.8/32"}
			accessRule.Spec.Include = []v1alpha1.CloudFlareAccessGroupRule{{
				IPRanges: ips},
			}
			for i := range ips {
				Expect(accessRule.ToCloudflare().Include[i]).To(Equal(cloudflare.AccessGroupIP{
					IP: struct {
						IP string "json:\"ip\""
					}{
						IP: ips[i],
					}}))
			}
		})

		It("can export serviceTokens to the cloudflare object", func() {
			ids := []v1alpha1.ServiceToken{{Value: "some_service_token"}, {Value: "some_other_service_token"}}
			accessRule.Spec.Include = []v1alpha1.CloudFlareAccessGroupRule{{
				ServiceToken: ids},
			}
			for i, id := range ids {
				Expect(accessRule.ToCloudflare().Include[i]).To(Equal(cloudflare.AccessGroupServiceToken{
					ServiceToken: struct {
						ID string "json:\"token_id\""
					}{
						ID: id.Value,
					},
				}))
			}
		})

		It("can export any serviceTokens to the cloudflare object", func() {
			validServiceToken := true
			accessRule.Spec.Include = []v1alpha1.CloudFlareAccessGroupRule{{
				AnyAccessServiceToken: &validServiceToken},
			}
			Expect(accessRule.ToCloudflare().Include[0]).To(Equal(cloudflare.AccessGroupAnyValidServiceToken{
				AnyValidServiceToken: struct{}{},
			}))
		})

		It("can export validCertificate to the cloudflare object", func() {
			validCert := true
			accessRule.Spec.Include = []v1alpha1.CloudFlareAccessGroupRule{{
				ValidCertificate: &validCert},
			}
			Expect(accessRule.ToCloudflare().Include[0]).To(Equal(cloudflare.AccessGroupCertificate{
				Certificate: struct{}{},
			}))
		})

		It("can export everyone to the cloudflare object", func() {
			everyone := true
			accessRule.Spec.Include = []v1alpha1.CloudFlareAccessGroupRule{{
				Everyone: &everyone},
			}
			Expect(accessRule.ToCloudflare().Include[0]).To(Equal(cloudflare.AccessGroupEveryone{
				Everyone: struct{}{},
			}))
		})

		It("can export Country to the cloudflare object", func() {
			accessRule.Spec.Include = []v1alpha1.CloudFlareAccessGroupRule{{
				Country: []string{"US", "UK"}},
			}
			for i, country := range accessRule.Spec.Include[0].Country {
				Expect(accessRule.ToCloudflare().Include[i]).To(Equal(cloudflare.AccessGroupGeo{
					Geo: struct {
						CountryCode string "json:\"country_code\""
					}{
						CountryCode: country,
					},
				}))
			}
		})

		It("can export accessGroups to the cloudflare object", func() {
			ids := []v1alpha1.AccessGroup{{Value: "first_access_group_id"}, {Value: "second_access_group_id"}}
			accessRule.Spec.Include = []v1alpha1.CloudFlareAccessGroupRule{{
				AccessGroups: ids},
			}
			for i, id := range ids {
				Expect(accessRule.ToCloudflare().Include[i]).To(Equal(cloudflare.AccessGroupAccessGroup{
					Group: struct {
						ID string "json:\"id\""
					}{
						ID: id.Value,
					},
				}))
			}
		})

		It("can export loginMethods to the cloudflare object", func() {
			accessRule.Spec.Include = []v1alpha1.CloudFlareAccessGroupRule{{
				LoginMethod: []string{"00000000-1234-5678-1234-123456789012"}},
			}

			for i, id := range accessRule.Spec.Include[0].LoginMethod {
				Expect(accessRule.ToCloudflare().Include[i]).To(Equal(cloudflare.AccessGroupLoginMethod{
					LoginMethod: struct {
						ID string "json:\"id\""
					}{
						ID: id,
					},
				}))
			}
		})
	})
})
