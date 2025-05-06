package cfcollections_test

import (
	"github.com/bojanzelic/cloudflare-zero-trust-operator/internal/cfcollections"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/cloudflare/cloudflare-go/v4/zero_trust"
)

var _ = Describe("AccessGroups", Label("AccessGroup"), func() {
	Context("AccessGroup test", func() {
		It("should be able able to find non-equality", func() {

			rule := &zero_trust.AccessRule{}
			rule.UnmarshalJSON([]byte(`{
				"email": {
					"email": "good@test.com"
				}
			}`))

			first := zero_trust.AccessGroupGetResponse{
				Name: "test",
				Include: []zero_trust.AccessRule{
					*rule,
				},
			}

			second := zero_trust.AccessGroupGetResponse{
				Name: "test",
				Include: []zero_trust.AccessRule{
					{
						Email: zero_trust.EmailRuleEmail{
							Email: "bad@test.com",
						},
					},
				},
			}

			Expect(cfcollections.AccessGroupEqual(first, second)).To(BeFalse())
		})
		It("should be able able to find equality", func() {
			rule := &zero_trust.AccessRule{}
			rule.UnmarshalJSON([]byte(`{
				"email": {
					"email": "test@test.com"
				}
			}`))

			first := zero_trust.AccessGroupGetResponse{
				Name: "test",
				Include: []zero_trust.AccessRule{
					*rule,
				},
			}

			second := zero_trust.AccessGroupGetResponse{
				Name: "test",
				Include: []zero_trust.AccessRule{
					{
						Email: zero_trust.EmailRuleEmail{
							Email: "test@test.com",
						},
					},
				},
			}

			Expect(cfcollections.AccessGroupEqual(first, second)).To(BeTrue())
		})

		It("should be able able to find equality with an array of elements", func() {

			rule1 := &zero_trust.AccessRule{}
			rule1.UnmarshalJSON([]byte(`{
				"email": {
					"email": "test@test.com"
				}
			}`))

			rule2 := &zero_trust.AccessRule{}
			rule2.UnmarshalJSON([]byte(`{
				"email": {
					"email": "test2@test.com"
				}
			}`))

			first := zero_trust.AccessGroupGetResponse{
				Name: "test",
				Include: []zero_trust.AccessRule{
					*rule1,
					*rule2,
				},
			}

			second := zero_trust.AccessGroupGetResponse{
				Name: "test",
				Include: []zero_trust.AccessRule{
					{
						Email: zero_trust.EmailRuleEmail{
							Email: "test@test.com",
						},
					},
					{
						Email: zero_trust.EmailRuleEmail{
							Email: "test2@test.com",
						},
					},
				},
			}

			Expect(cfcollections.AccessGroupEqual(first, second)).To(BeTrue())
		})
	})
	Context("AccessGroupCollection test", func() {
		It("Should be able to find by name", func() {
			groups := cfcollections.AccessGroupCollection{
				{
					Name: "first",
					Include: []zero_trust.AccessRule{
						{
							Email: zero_trust.EmailRuleEmail{
								Email: "test@test.com",
							},
						},
					},
				},
				{
					Name: "second",
					Include: []zero_trust.AccessRule{
						{
							Email: zero_trust.EmailRuleEmail{
								Email: "test2@test.com",
							},
						},
					},
				},
			}

			Expect(groups.GetByName("first")).To(Equal(&groups[0]))
			Expect(groups.GetByName("second")).To(Equal(&groups[1]))
		})
	})
})
