package cfcollections_test

import (
	"github.com/bojanzelic/cloudflare-zero-trust-operator/internal/cfcollections"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	cloudflare "github.com/cloudflare/cloudflare-go/v4"
)

var _ = Describe("AccessGroups", Label("AccessGroup"), func() {
	Context("AccessGroup test", func() {
		It("should be able able to find non-equality", func() {
			first := cloudflare.AccessGroup{
				Name: "test",
				Include: []interface{}{
					map[string]interface{}{
						"email": map[string]interface{}{
							"email": "good@test.com",
						},
					},
				},
			}

			second := cloudflare.AccessGroup{
				Name: "test",
				Include: []interface{}{cloudflare.AccessGroupEmail{
					Email: struct {
						Email string "json:\"email\""
					}{
						Email: "bad@test.com",
					},
				}},
			}

			Expect(cfcollections.AccessGroupEqual(first, second)).To(BeFalse())
		})
		It("should be able able to find equality", func() {
			first := cloudflare.AccessGroup{
				Name: "test",
				Include: []interface{}{
					map[string]interface{}{
						"email": map[string]interface{}{
							"email": "test@test.com",
						},
					},
				},
			}

			second := cloudflare.AccessGroup{
				Name: "test",
				Include: []interface{}{cloudflare.AccessGroupEmail{
					Email: struct {
						Email string "json:\"email\""
					}{
						Email: "test@test.com",
					},
				}},
			}

			Expect(cfcollections.AccessGroupEqual(first, second)).To(BeTrue())
		})

		It("should be able able to find equality with an array of elements", func() {
			first := cloudflare.AccessGroup{
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

			second := cloudflare.AccessGroup{
				Name: "test",
				Include: []interface{}{
					cloudflare.AccessGroupEmail{
						Email: struct {
							Email string "json:\"email\""
						}{
							Email: "test@test.com",
						},
					},
					cloudflare.AccessGroupEmail{
						Email: struct {
							Email string "json:\"email\""
						}{
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
