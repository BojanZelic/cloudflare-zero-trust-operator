package cfapi

//
// These API call are most probably used only for tests
//

import (
	"context"

	cloudflare "github.com/cloudflare/cloudflare-go/v4"
	"github.com/cloudflare/cloudflare-go/v4/zero_trust"
	"github.com/pkg/errors"
)

func (a *API) IdentityProviders(ctx context.Context) (*[]zero_trust.IdentityProviderListResponse, error) {
	//
	iter := a.client.ZeroTrust.IdentityProviders.ListAutoPaging(ctx, zero_trust.IdentityProviderListParams{
		AccountID: cloudflare.F(a.CFAccountID),
	})

	idProviders := []zero_trust.IdentityProviderListResponse{}
	for iter.Next() {
		idProviders = append(idProviders, iter.Current())
	}

	//
	return &idProviders, errors.Wrap(iter.Err(), "unable to get identity providers")
}
