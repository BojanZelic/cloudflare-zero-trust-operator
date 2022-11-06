package v1alpha1_test

import (
	"testing"

	"github.com/bojanzelic/cloudflare-zero-trust-operator/api/v1alpha1"
	cloudflare "github.com/cloudflare/cloudflare-go"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	//+kubebuilder:scaffold:imports
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
			ids := []string{"some_service_token", "some_other_service_token"}
			accessRule.Spec.Include = []v1alpha1.CloudFlareAccessGroupRule{{
				ServiceToken: ids},
			}
			for i, id := range ids {
				Expect(accessRule.ToCloudflare().Include[i]).To(Equal(cloudflare.AccessGroupServiceToken{
					ServiceToken: struct {
						ID string "json:\"token_id\""
					}{
						ID: id,
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

		It("can export accessGroups to the cloudflare object", func() {
			ids := []string{"first_access_group_id", "second_access_group_id"}
			accessRule.Spec.Include = []v1alpha1.CloudFlareAccessGroupRule{{
				AccessGroups: ids},
			}
			for i, id := range ids {
				Expect(accessRule.ToCloudflare().Include[i]).To(Equal(cloudflare.AccessGroupAccessGroup{
					Group: struct {
						ID string "json:\"id\""
					}{
						ID: id,
					},
				}))
			}
		})
	})
})
