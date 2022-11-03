package cfapi

import (
	"context"
	"encoding/json"
	"reflect"

	"github.com/bojanzelic/cloudflare-zero-trust-operator/internal/config"
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

func NewAccessGroupEmail(email string) cloudflare.AccessGroupEmail {
	return cloudflare.AccessGroupEmail{Email: struct {
		Email string "json:\"email\""
	}{
		Email: email,
	}}
}

func AcessGroupEmailEqual(first cloudflare.AccessGroup, second cloudflare.AccessGroup) bool {
	v1, _ := json.Marshal(first.Include)
	v2, _ := json.Marshal(second.Include)
	if !reflect.DeepEqual(v1, v2) {
		return false
	}
	v1, _ = json.Marshal(first.Exclude)
	v2, _ = json.Marshal(second.Exclude)
	if !reflect.DeepEqual(v1, v2) {
		return false
	}
	v1, _ = json.Marshal(first.Require)
	v2, _ = json.Marshal(second.Require)
	if !reflect.DeepEqual(v1, v2) {
		return false
	}

	return true
}

func (a *API) AccessGroups(ctx context.Context) ([]cloudflare.AccessGroup, error) {
	cfAccessGroups, _, err := a.client.AccessGroups(ctx, config.CLOUDFLARE_ACCOUNT_ID, cloudflare.PaginationOptions{})

	// for i, ag := range cfAccessGroups {
	// 	cfAccessGroups[i].Include

	// }

	return cfAccessGroups, errors.Wrap(err, "unable to get access groups")
}

func (a *API) AccessGroup(ctx context.Context, accessGroupID string) (cloudflare.AccessGroup, error) {
	cfAG, err := a.client.AccessGroup(ctx, config.CLOUDFLARE_ACCOUNT_ID, accessGroupID)
	return cfAG, errors.Wrap(err, "unable to get access group")
}

func (a *API) CreateAccessGroup(ctx context.Context, ag cloudflare.AccessGroup) (cloudflare.AccessGroup, error) {
	cfAG, err := a.client.CreateAccessGroup(ctx, config.CLOUDFLARE_ACCOUNT_ID, ag)
	return cfAG, errors.Wrap(err, "unable to create access groups")
}

func (a *API) UpdateAccessGroup(ctx context.Context, ag cloudflare.AccessGroup) (cloudflare.AccessGroup, error) {
	cfAG, err := a.client.UpdateAccessGroup(ctx, config.CLOUDFLARE_ACCOUNT_ID, ag)
	return cfAG, errors.Wrap(err, "unable to update access groups")
}
