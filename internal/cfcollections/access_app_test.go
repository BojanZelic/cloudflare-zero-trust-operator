package cfcollections_test

import (
	"github.com/bojanzelic/cloudflare-zero-trust-operator/internal/cfcollections"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/cloudflare/cloudflare-go/v4/zero_trust"
)

var _ = Describe("AccessApp", Label("AccessApp"), func() {
	Context("AccessApp test", func() {
		It("Empty Apps should be equal", func() {
			first := zero_trust.AccessApplicationGetResponse{}
			second := zero_trust.AccessApplicationGetResponse{}

			Expect(cfcollections.AccessAppEqual(first, second)).To(BeTrue())
		})

		It("Differences should no be equal", func() {
			first := zero_trust.AccessApplicationGetResponse{
				Name: "first",
			}
			second := zero_trust.AccessApplicationGetResponse{
				Name: "second",
			}

			Expect(cfcollections.AccessAppEqual(first, second)).To(BeFalse())
		})
	})
})
