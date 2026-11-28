package cfcollections_test

import (
	"github.com/bojanzelic/cloudflare-zero-trust-operator/internal/cfapi"
	"github.com/bojanzelic/cloudflare-zero-trust-operator/internal/cfcollections"
	"github.com/bojanzelic/cloudflare-zero-trust-operator/internal/cftypes"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("AccessPolicy", Label("AccessPolicy"), func() {
	Context("AccessPolicy test", func() {
		It("should be able to determine equality", func() {
			first := cftypes.AccessPolicy{
				Name:       "test",
				Precedence: 1,
				Include: []interface{}{
					map[string]interface{}{
						"email": map[string]interface{}{
							"email": "test@test.com",
						},
					},
				},
			}

			second := cftypes.AccessPolicy{
				Name:       "test",
				Precedence: 1,
				Include: []interface{}{
					cfapi.NewAccessGroupEmail("test@test.com"),
				},
			}

			Expect(cfcollections.AccessPoliciesEqual(&first, &second)).To(BeTrue())
		})
	})
	Context("AccessPolicyCollection test", func() {
		It("Should be able to sort by precidence", func() {
			aps := cfcollections.AccessPolicyCollection{
				{
					Name:       "test4",
					Precedence: 4,
				},
				{
					Name:       "test3",
					Precedence: 3,
				},
				{
					Name:       "test2",
					Precedence: 2,
				},
				{
					Name:       "test1",
					Precedence: 1,
				},
				{
					Name:       "test5",
					Precedence: 5,
				},
			}

			aps.SortByPrecidence()

			prevAP := cftypes.AccessPolicy{Precedence: 0}
			for _, ap := range aps {
				Expect(ap.Precedence > prevAP.Precedence).To(BeTrue())
				prevAP = ap
			}
		})
	})
})
