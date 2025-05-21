package cfapi

import (
	"context"

	"github.com/bojanzelic/cloudflare-zero-trust-operator/internal/config"
	cloudflare "github.com/cloudflare/cloudflare-go/v4"
	"github.com/cloudflare/cloudflare-go/v4/option"
)

type API struct {
	CFAccountID    string
	client         *cloudflare.Client
	optionalTracer *CloudflareResourceCreationTracer
	ctx            context.Context
}

func FromConfig(ctx context.Context, config config.ZeroTrustConfig, optionalTracer *CloudflareResourceCreationTracer) *API {
	return New(ctx,
		config.APIToken,
		config.APIKey,
		config.APIEmail,
		config.AccountID,
		optionalTracer,
	)
}

func New(
	ctx context.Context,
	cfAPIToken string,
	cfAPIKey string,
	cfAPIEmail string,
	cfAccountID string,
	optionalTracer *CloudflareResourceCreationTracer,
) *API {
	//
	var api *cloudflare.Client
	if cfAPIToken != "" {
		api = cloudflare.NewClient(option.WithAPIToken(cfAPIToken))
	} else {
		api = cloudflare.NewClient(option.WithAPIKey(cfAPIKey), option.WithAPIEmail(cfAPIEmail))
	}

	//
	return &API{
		ctx: ctx,
		//
		CFAccountID: cfAccountID,
		client:      api,
		//
		optionalTracer: optionalTracer,
	}
}
