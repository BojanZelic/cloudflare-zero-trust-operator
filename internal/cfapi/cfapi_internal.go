package cfapi

//
// These API call are most probably used only for tests
//

import (
	"context"

	"github.com/Southclaws/fault"
	"github.com/Southclaws/fault/fmsg"
	cloudflare "github.com/cloudflare/cloudflare-go/v4"
	"github.com/cloudflare/cloudflare-go/v4/zero_trust"
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
	return &idProviders, fault.Wrap(iter.Err(), fmsg.With("unable to get identity providers"))
}
