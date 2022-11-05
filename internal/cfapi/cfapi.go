package cfapi

import (
	"context"

	cloudflare "github.com/cloudflare/cloudflare-go"
	"github.com/pkg/errors"
)

type API struct {
	CFAccountID string
	client      *cloudflare.API
}

func New(CFApiToken string, CFAPIKey string, CFAPIEmail string, CFAccountID string) (*API, error) {
	var err error
	var api *cloudflare.API

	if CFApiToken != "" {
		api, err = cloudflare.NewWithAPIToken(CFApiToken)
	} else {
		api, err = cloudflare.New(CFAPIKey, CFAPIEmail)
	}

	return &API{
		CFAccountID: CFAccountID,
		client:      api,
	}, err
}

func (a *API) AccessGroups(ctx context.Context) ([]cloudflare.AccessGroup, error) {
	cfAccessGroups, _, err := a.client.AccessGroups(ctx, a.CFAccountID, cloudflare.PaginationOptions{})
	return cfAccessGroups, errors.Wrap(err, "unable to get access groups")
}

func (a *API) AccessGroup(ctx context.Context, accessGroupID string) (cloudflare.AccessGroup, error) {
	cfAG, err := a.client.AccessGroup(ctx, a.CFAccountID, accessGroupID)
	return cfAG, errors.Wrap(err, "unable to get access group")
}

func (a *API) CreateAccessGroup(ctx context.Context, ag cloudflare.AccessGroup) (cloudflare.AccessGroup, error) {
	cfAG, err := a.client.CreateAccessGroup(ctx, a.CFAccountID, ag)
	return cfAG, errors.Wrap(err, "unable to create access groups")
}

func (a *API) UpdateAccessGroup(ctx context.Context, ag cloudflare.AccessGroup) (cloudflare.AccessGroup, error) {
	cfAG, err := a.client.UpdateAccessGroup(ctx, a.CFAccountID, ag)
	return cfAG, errors.Wrap(err, "unable to update access groups")
}
