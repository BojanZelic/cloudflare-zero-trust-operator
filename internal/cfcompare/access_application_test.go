package cfcompare_test

import (
	"context"

	"github.com/bojanzelic/cloudflare-zero-trust-operator/api/v4alpha1"
	"github.com/bojanzelic/cloudflare-zero-trust-operator/internal/cfcompare"
	"github.com/bojanzelic/cloudflare-zero-trust-operator/internal/logger"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/zap/zapcore"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/cloudflare/cloudflare-go/v4/zero_trust"
)

var _ = Describe("AccessApp", Label("AccessApp"), func() {
	Context("AccessApp test", func() {
		//
		ctx := context.Background()
		log := logger.NewFaultLogger(
			zap.New(
				zap.WriteTo(GinkgoWriter),
				zap.UseDevMode(true),
				zap.StacktraceLevel(zapcore.DPanicLevel), // only print stacktraces for panics and fatal
			),
			&logger.FaultLoggerOptions{
				DismissErrorVerbose: true,
			},
		)

		It("Empty Apps should be equal", func() {
			first := zero_trust.AccessApplicationGetResponse{}
			second := v4alpha1.CloudflareAccessApplication{}
			Expect(cfcompare.AreAccessApplicationsEquivalent(ctx, &log, &first, &second)).To(BeTrue())
		})

		It("Default Apps should be equal", func() {
			first := zero_trust.AccessApplicationGetResponse{
				Type: string(zero_trust.ApplicationTypeSelfHosted),
			}
			second := v4alpha1.CloudflareAccessApplication{
				Spec: v4alpha1.CloudflareAccessApplicationSpec{
					Type: string(zero_trust.ApplicationTypeSelfHosted),
				},
			}
			Expect(cfcompare.AreAccessApplicationsEquivalent(ctx, &log, &first, &second)).To(BeTrue())
		})

		It("Differences on names in self_hosted should no be equal", func() {
			first := zero_trust.AccessApplicationGetResponse{
				Type: string(zero_trust.ApplicationTypeSelfHosted),
				Name: "first",
			}
			second := v4alpha1.CloudflareAccessApplication{
				Spec: v4alpha1.CloudflareAccessApplicationSpec{
					Type: string(zero_trust.ApplicationTypeSelfHosted),
					Name: "second",
				},
			}

			Expect(cfcompare.AreAccessApplicationsEquivalent(ctx, &log, &first, &second)).To(BeFalse())
		})

		It("Differences on names in warp should be equal", func() {
			first := zero_trust.AccessApplicationGetResponse{
				Type: string(zero_trust.ApplicationTypeWARP),
				Name: "first",
			}
			second := v4alpha1.CloudflareAccessApplication{
				Spec: v4alpha1.CloudflareAccessApplicationSpec{
					Type: string(zero_trust.ApplicationTypeWARP),
					Name: "second",
				},
			}

			Expect(cfcompare.AreAccessApplicationsEquivalent(ctx, &log, &first, &second)).To(BeTrue())
		})
	})
})
