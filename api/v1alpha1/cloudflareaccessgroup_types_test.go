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
				// Expect(accessRule.ToCloudflare().Require[i]).To(Equal(cloudflare.AccessGroupEmail{
				// 	Email: struct {
				// 		Email string "json:\"email\""
				// 	}{
				// 		Email: emails[i],
				// 	},
				// }))
				// Expect(accessRule.ToCloudflare().Exclude[i]).To(Equal(cloudflare.AccessGroupEmail{
				// 	Email: struct {
				// 		Email string "json:\"email\""
				// 	}{
				// 		Email: emails[i],
				// 	},
				// }))
			}
		})

		It("can export Required emails to the cloudflare object", func() {
			emails := []string{"test@email.com", "test2@email.com"}
			accessRule.Spec.Include = []v1alpha1.CloudFlareAccessGroupRule{{
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
			}
		})
	})
})
