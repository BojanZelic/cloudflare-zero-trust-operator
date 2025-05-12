package cfcompare_test

import (
	"github.com/bojanzelic/cloudflare-zero-trust-operator/api/v4alpha1"
	"github.com/bojanzelic/cloudflare-zero-trust-operator/internal/cfcompare"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/cloudflare/cloudflare-go/v4/zero_trust"
)

var _ = Describe("AccessGroups", Label("AccessGroup"), func() {
	Context("AccessGroup test", func() {
		It("should be able able to find non-equality", func() {

			rule := &zero_trust.AccessRule{}
			err := rule.UnmarshalJSON([]byte(`{
				"email": {
					"email": "good@test.com"
				}
			}`))

			Expect(err).NotTo(HaveOccurred())

			first := zero_trust.AccessGroupGetResponse{
				Name: "test",
				Include: []zero_trust.AccessRule{
					*rule,
				},
			}

			second := v4alpha1.CloudflareAccessGroup{
				Spec: v4alpha1.CloudflareAccessGroupSpec{
					Name: "test",
					Include: v4alpha1.CloudFlareAccessRules{
						Emails: []string{"bad@test.com"},
					},
				},
			}

			Expect(cfcompare.AreAccessGroupsEquivalent(&first, &second)).To(BeFalse())
		})
		It("should be able able to find equality", func() {
			rule := &zero_trust.AccessRule{}
			err := rule.UnmarshalJSON([]byte(`{
				"email": {
					"email": "test@test.com"
				}
			}`))

			Expect(err).NotTo(HaveOccurred())

			first := zero_trust.AccessGroupGetResponse{
				Name: "test",
				Include: []zero_trust.AccessRule{
					*rule,
				},
			}

			second := v4alpha1.CloudflareAccessGroup{
				Spec: v4alpha1.CloudflareAccessGroupSpec{
					Name: "test",
					Include: v4alpha1.CloudFlareAccessRules{
						Emails: []string{"test@test.com"},
					},
				},
			}

			Expect(cfcompare.AreAccessGroupsEquivalent(&first, &second)).To(BeTrue())
		})

		It("should be able able to find equality with an array of elements", func() {

			rule1 := &zero_trust.AccessRule{}
			err := rule1.UnmarshalJSON([]byte(`{
				"email": {
					"email": "test1@test.com"
				}
			}`))

			Expect(err).NotTo(HaveOccurred())

			rule2 := &zero_trust.AccessRule{}
			err = rule2.UnmarshalJSON([]byte(`{
				"email": {
					"email": "test2@test.com"
				}
			}`))

			Expect(err).NotTo(HaveOccurred())

			first := zero_trust.AccessGroupGetResponse{
				Name: "test",
				Include: []zero_trust.AccessRule{
					*rule1,
					*rule2,
				},
			}

			second := v4alpha1.CloudflareAccessGroup{
				Spec: v4alpha1.CloudflareAccessGroupSpec{
					Name: "test",
					Include: v4alpha1.CloudFlareAccessRules{
						Emails: []string{"test1@test.com", "test2@test.com"},
					},
				},
			}

			Expect(cfcompare.AreAccessGroupsEquivalent(&first, &second)).To(BeTrue())
		})
	})
})
