package cfcollections_test

import (
	"github.com/bojanzelic/cloudflare-zero-trust-operator/internal/cfcollections"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	cloudflare "github.com/cloudflare/cloudflare-go"
)

var _ = Describe("AccessGroups", Label("AccessGroup"), func() {
	Context("AccessGroupCollection test", func() {
		It("Should be able to find by name", func() {
			groups := cfcollections.AccessGroupCollection{
				{
					Name: "first",
					Include: []interface{}{
						cloudflare.AccessGroupEmail{
							Email: struct {
								Email string "json:\"email\""
							}{
								Email: "test@test.com",
							},
						},
					},
				},
				{
					Name: "second",
					Include: []interface{}{
						cloudflare.AccessGroupEmail{
							Email: struct {
								Email string "json:\"email\""
							}{
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
