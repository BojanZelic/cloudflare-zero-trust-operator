package cfcollections_test

import (
	"github.com/bojanzelic/cloudflare-zero-trust-operator/internal/cfcollections"
	"github.com/bojanzelic/cloudflare-zero-trust-operator/internal/cftypes"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("AccessApp", Label("AccessApp"), func() {
	Context("AccessApp test", func() {
		It("Empty Apps should be equal", func() {
			first := cftypes.AccessApplication{}
			second := cftypes.AccessApplication{}

			Expect(cfcollections.AccessAppEqual(first, second)).To(BeTrue())
		})

		It("Differences should no be equal", func() {
			first := cftypes.AccessApplication{
				Name: "first",
			}
			second := cftypes.AccessApplication{
				Name: "second",
			}

			Expect(cfcollections.AccessAppEqual(first, second)).To(BeFalse())
		})
	})
})
