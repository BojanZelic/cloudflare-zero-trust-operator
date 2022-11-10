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

func (a *API) AccessApplications(ctx context.Context) ([]cloudflare.AccessApplication, error) {
	apps, _, err := a.client.AccessApplications(ctx, a.CFAccountID, cloudflare.PaginationOptions{})
	return apps, errors.Wrap(err, "unable to get access applications")
}

func (a *API) FindAccessApplicationByDomain(ctx context.Context, domain string) (*cloudflare.AccessApplication, error) {
	apps, _, err := a.client.AccessApplications(ctx, a.CFAccountID, cloudflare.PaginationOptions{})

	if err != nil {
		return nil, errors.Wrap(err, "unable to get access applications")
	}

	for _, g := range apps {
		if g.Domain == domain {
			return &g, nil
		}
	}

	return nil, nil
}

func (a *API) AccessApplication(ctx context.Context, AccessApplicationID string) (cloudflare.AccessApplication, error) {
	cfAG, err := a.client.AccessApplication(ctx, a.CFAccountID, AccessApplicationID)
	return cfAG, errors.Wrap(err, "unable to get access application")
}

func (a *API) CreateAccessApplication(ctx context.Context, ag cloudflare.AccessApplication) (cloudflare.AccessApplication, error) {
	cfAG, err := a.client.CreateAccessApplication(ctx, a.CFAccountID, ag)
	return cfAG, errors.Wrap(err, "unable to create access applications")
}

func (a *API) UpdateAccessApplication(ctx context.Context, ag cloudflare.AccessApplication) (cloudflare.AccessApplication, error) {
	cfAG, err := a.client.UpdateAccessApplication(ctx, a.CFAccountID, ag)
	return cfAG, errors.Wrap(err, "unable to update access applications")
}

func (a *API) AccessPolicies(ctx context.Context, appID string) ([]cloudflare.AccessPolicy, error) {
	apps, _, err := a.client.AccessPolicies(ctx, a.CFAccountID, appID, cloudflare.PaginationOptions{})
	return apps, errors.Wrap(err, "unable to get access Policies")
}

func (a *API) CreateAccessPolicy(ctx context.Context, appID string, ag cloudflare.AccessPolicy) (cloudflare.AccessPolicy, error) {
	cfAG, err := a.client.CreateAccessPolicy(ctx, a.CFAccountID, appID, ag)
	return cfAG, errors.Wrap(err, "unable to create access Policy")
}

func (a *API) UpdateAccessPolicy(ctx context.Context, appID string, ag cloudflare.AccessPolicy) (cloudflare.AccessPolicy, error) {
	cfAG, err := a.client.UpdateAccessPolicy(ctx, a.CFAccountID, appID, ag)
	return cfAG, errors.Wrap(err, "unable to update access Policy")
}

func (a *API) DeleteAccessPolicy(ctx context.Context, appID string, policyID string) error {
	err := a.client.DeleteAccessPolicy(ctx, a.CFAccountID, appID, policyID)
	return errors.Wrap(err, "unable to update access Policy")
}
