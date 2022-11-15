package config_test

import (
	"os"
	"testing"

	"github.com/bojanzelic/cloudflare-zero-trust-operator/internal/config"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestBooks(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "CloudflareAccessGroup Suite")
}

var _ = Describe("Config", Label("Config"), func() {
	Context("Config object test", func() {
		BeforeEach(func() {
			os.Setenv("CLOUDFLARE_API_EMAIL", "test@test.com")
			os.Setenv("CLOUDFLARE_API_KEY", "1123457890")
			os.Setenv("CLOUDFLARE_API_TOKEN", "2123457890")
			os.Setenv("CLOUDFLARE_ACCOUNT_ID", "3123457890")
		})

		AfterEach(func() {
			os.Unsetenv("CLOUDFLARE_API_EMAIL")
			os.Unsetenv("CLOUDFLARE_API_KEY")
			os.Unsetenv("CLOUDFLARE_API_TOKEN")
			os.Unsetenv("CLOUDFLARE_ACCOUNT_ID")
		})

		It("Should load ENV values automatically", func() {
			config.SetConfigDefaults()
			ztConfig := config.ParseCloudflareConfig(&v1.ObjectMeta{})
			Expect(ztConfig.IsValid()).To(Equal(true))
			Expect(ztConfig.APIEmail).To(Equal("test@test.com"))
			Expect(ztConfig.APIKey).To(Equal("1123457890"))
			Expect(ztConfig.APIToken).To(Equal("2123457890"))
			Expect(ztConfig.AccountID).To(Equal("3123457890"))
		})
	})
})
