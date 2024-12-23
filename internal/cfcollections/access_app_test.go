package cfcollections_test

import (
	"github.com/kadaan/cloudflare-zero-trust-operator/internal/cfcollections"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	cloudflare "github.com/cloudflare/cloudflare-go"
)

var _ = Describe("AccessApp", Label("AccessApp"), func() {
	Context("AccessApp test", func() {
		It("Empty Apps should be equal", func() {
			first := cloudflare.AccessApplication{}
			second := cloudflare.AccessApplication{}

			Expect(cfcollections.AccessAppEqual(first, second)).To(BeTrue())
		})

		It("Differences should no be equal", func() {
			first := cloudflare.AccessApplication{
				Name: "first",
			}
			second := cloudflare.AccessApplication{
				Name: "second",
			}

			Expect(cfcollections.AccessAppEqual(first, second)).To(BeFalse())
		})
	})
})
