package config

import (
	"errors"

	"github.com/spf13/viper"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ZeroTrustConfig struct {
	APIEmail  string
	APIKey    string
	APIToken  string
	AccountID string
}

var (
	ErrMissingCFFields  = errors.New("missing one of CLOUDFLARE_API_TOKEN or (CLOUDFLARE_API_EMAIL and CLOUDFLARE_API_KEY) needs to be set")
	ErrMissingAccountID = errors.New("missing CLOUDFLARE_ACCOUNT_ID needs to be set")
)

func SetConfigDefaults() {
	viper.SetDefault("cloudflare_api_email", "")
	viper.SetDefault("cloudflare_api_key", "")
	viper.SetDefault("cloudflare_api_token", "")
	viper.SetDefault("cloudflare_account_id", "")
	viper.AutomaticEnv()
}

func ParseCloudflareConfig(obj metav1.Object) ZeroTrustConfig {
	cloudflareConfig := ZeroTrustConfig{}

	annotations := obj.GetAnnotations()

	cloudflareConfig.AccountID = viper.GetString("cloudflare_account_id")
	cloudflareConfig.APIEmail = viper.GetString("cloudflare_api_email")
	cloudflareConfig.APIToken = viper.GetString("cloudflare_api_token")
	cloudflareConfig.APIKey = viper.GetString("cloudflare_api_key")

	if val, ok := annotations["cloudflare.zero-trust.kadaan.info/account_id"]; ok {
		cloudflareConfig.AccountID = val
	} else {
		cloudflareConfig.AccountID = viper.GetString("cloudflare_account_id")
	}

	return cloudflareConfig
}

func (c ZeroTrustConfig) IsValid() (bool, error) {
	if c.AccountID == "" {
		return false, ErrMissingAccountID
	}

	if c.APIToken == "" && (c.APIEmail == "" && c.APIKey == "") {
		return false, ErrMissingCFFields
	}

	return true, nil
}
