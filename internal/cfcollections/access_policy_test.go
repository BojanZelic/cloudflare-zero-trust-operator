package cfcollections_test

import (
	"github.com/bojanzelic/cloudflare-zero-trust-operator/internal/cfcollections"
	"github.com/cloudflare/cloudflare-go/v4/zero_trust"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("AccessPolicy", Label("AccessPolicy"), func() {
	Context("AccessPolicy test", func() {
		It("should be able to determine equality", func() {

			rule := &zero_trust.AccessRule{}
			rule.UnmarshalJSON([]byte(`{
				"email": {
					"email": "test@test.com"
				}
			}`))

			first := zero_trust.AccessApplicationPolicyListResponse{
				ApplicationPolicy: zero_trust.ApplicationPolicy{
					Name: "test",
					Include: []zero_trust.AccessRule{
						*rule,
					},
				},
				Precedence: 1,
			}

			second := zero_trust.AccessApplicationPolicyListResponse{
				ApplicationPolicy: zero_trust.ApplicationPolicy{
					Name: "test",
					Include: []zero_trust.AccessRule{
						{
							Email: zero_trust.EmailRuleEmail{
								Email: "test@test.com",
							},
						},
					},
				},
				Precedence: 1,
			}

			Expect(cfcollections.AccessPoliciesEqual(&first, &second)).To(BeTrue())
		})
	})
	Context("LegacyAccessPolicyCollection test", func() {
		It("Should be able to sort by precidence", func() {
			aps := cfcollections.LegacyAccessPolicyCollection{
				{
					ApplicationPolicy: zero_trust.ApplicationPolicy{
						Name: "test4",
					},
					Precedence: 4,
				},
				{
					ApplicationPolicy: zero_trust.ApplicationPolicy{
						Name: "test3",
					},
					Precedence: 3,
				},
				{
					ApplicationPolicy: zero_trust.ApplicationPolicy{
						Name: "test2",
					},
					Precedence: 2,
				},
				{
					ApplicationPolicy: zero_trust.ApplicationPolicy{
						Name: "test1",
					},
					Precedence: 1,
				},
				{
					ApplicationPolicy: zero_trust.ApplicationPolicy{
						Name: "test5",
					},
					Precedence: 5,
				},
			}

			aps.SortByPrecedence()

			prevAP := zero_trust.AccessApplicationPolicyListResponse{Precedence: 0}
			for _, ap := range aps {
				Expect(ap.Precedence > prevAP.Precedence).To(BeTrue())
				prevAP = ap
			}
		})
	})
})
