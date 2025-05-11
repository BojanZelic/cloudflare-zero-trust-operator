package cfcompare_test

import (
	"github.com/bojanzelic/cloudflare-zero-trust-operator/api/v4alpha1"
	"github.com/bojanzelic/cloudflare-zero-trust-operator/internal/cfcompare"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/cloudflare/cloudflare-go/v4/zero_trust"
)

var _ = Describe("AccessApp", Label("AccessApp"), func() {
	Context("AccessApp test", func() {
		It("Empty Apps should be equal", func() {
			first := zero_trust.AccessApplicationGetResponse{}
			second := v4alpha1.CloudflareAccessApplication{}

			Expect(cfcompare.AreAccessApplicationsEquivalent(&first, &second)).To(BeTrue())
		})

		It("Differences should no be equal", func() {
			first := zero_trust.AccessApplicationGetResponse{
				Name: "first",
			}
			second := v4alpha1.CloudflareAccessApplication{
				Spec: v4alpha1.CloudflareAccessApplicationSpec{
					Name: "second",
				},
			}

			Expect(cfcompare.AreAccessApplicationsEquivalent(&first, &second)).To(BeFalse())
		})
	})
})
