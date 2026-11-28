package cfcollections_test

import (
	"github.com/bojanzelic/cloudflare-zero-trust-operator/internal/cfapi"
	"github.com/bojanzelic/cloudflare-zero-trust-operator/internal/cfcollections"
	"github.com/bojanzelic/cloudflare-zero-trust-operator/internal/cftypes"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("AccessGroups", Label("AccessGroup"), func() {
	Context("AccessGroup test", func() {
		It("should be able able to find non-equality", func() {
			first := cftypes.AccessGroup{
				Name: "test",
				Include: []interface{}{
					map[string]interface{}{
						"email": map[string]interface{}{
							"email": "good@test.com",
						},
					},
				},
			}

			second := cftypes.AccessGroup{
				Name: "test",
				Include: []interface{}{
					cfapi.NewAccessGroupEmail("bad@test.com"),
				},
			}

			Expect(cfcollections.AccessGroupEqual(first, second)).To(BeFalse())
		})
		It("should be able able to find equality", func() {
			first := cftypes.AccessGroup{
				Name: "test",
				Include: []interface{}{
					map[string]interface{}{
						"email": map[string]interface{}{
							"email": "test@test.com",
						},
					},
				},
			}

			second := cftypes.AccessGroup{
				Name: "test",
				Include: []interface{}{
					cfapi.NewAccessGroupEmail("test@test.com"),
				},
			}

			Expect(cfcollections.AccessGroupEqual(first, second)).To(BeTrue())
		})

		It("should be able able to find equality with an array of elements", func() {
			first := cftypes.AccessGroup{
				Name: "test",
				Include: []interface{}{
					map[string]interface{}{
						"email": map[string]interface{}{
							"email": "test@test.com",
						},
					},
					map[string]interface{}{
						"email": map[string]interface{}{
							"email": "test2@test.com",
						},
					},
				},
			}

			second := cftypes.AccessGroup{
				Name: "test",
				Include: []interface{}{
					cfapi.NewAccessGroupEmail("test@test.com"),
					cfapi.NewAccessGroupEmail("test2@test.com"),
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
					Include: []interface{}{
						cfapi.NewAccessGroupEmail("test@test.com"),
					},
				},
				{
					Name: "second",
					Include: []interface{}{
						cfapi.NewAccessGroupEmail("test2@test.com"),
					},
				},
			}

			Expect(groups.GetByName("first")).To(Equal(&groups[0]))
			Expect(groups.GetByName("second")).To(Equal(&groups[1]))
		})
	})
})
