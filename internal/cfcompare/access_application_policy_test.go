package cfcompare_test

import (
	"github.com/bojanzelic/cloudflare-zero-trust-operator/api/v4alpha1"
	"github.com/bojanzelic/cloudflare-zero-trust-operator/internal/cfcompare"
	"github.com/cloudflare/cloudflare-go/v4/zero_trust"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("AccessApplicationPolicy", Label("AccessApplicationPolicy"), func() {
	Context("AccessApplicationPolicy test", func() {
		It("should be able to determine equality", func() {

			rule := &zero_trust.AccessRule{}
			err := rule.UnmarshalJSON([]byte(`{
				"email": {
					"email": "test@test.com"
				}
			}`))

			Expect(err).NotTo(HaveOccurred())

			first := zero_trust.AccessApplicationGetResponse{
				Policies: []map[string]any{{
					"name":       "test",
					"id":         "1232313123123",
					"precedence": 1,
					"include": map[string]any{
						"email": map[string]any{
							"email": "test@test.com",
						},
					},
				}},
			}

			second := v4alpha1.CloudflareAccessApplication{
				Status: v4alpha1.CloudflareAccessApplicationStatus{
					ReusablePolicyIDs: []string{"1232313123123"},
				},
			}

			Expect(cfcompare.DoK8SAccessPoliciesMatch(&first, &second)).To(BeTrue())
		})
	})
})
