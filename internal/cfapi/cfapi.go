package cfapi

import (
	"context"

	"github.com/bojanzelic/cloudflare-zero-trust-operator/internal/cfcollections"
	"github.com/bojanzelic/cloudflare-zero-trust-operator/internal/cftypes"
	cloudflare "github.com/cloudflare/cloudflare-go"
	"github.com/pkg/errors"
)

type API struct {
	CFAccountID string
	client      *cloudflare.API
}

func New(cfAPIToken string, cfAPIKey string, cfAPIEmail string, cfAccountID string) (*API, error) {
	var err error
	var api *cloudflare.API

	if cfAPIToken != "" {
		api, err = cloudflare.NewWithAPIToken(cfAPIToken)
	} else {
		api, err = cloudflare.New(cfAPIKey, cfAPIEmail)
	}

	return &API{
		CFAccountID: cfAccountID,
		client:      api,
	}, errors.Wrap(err, "error initializing Cloudflare API")
}

func (a *API) AccessGroups(ctx context.Context) (cfcollections.AccessGroupCollection, error) {
	cfAccessGroups, _, err := a.client.AccessGroups(ctx, a.CFAccountID, cloudflare.PaginationOptions{})
	cfAccessGroupCollection := cfcollections.AccessGroupCollection(cfAccessGroups)

	return cfAccessGroupCollection, errors.Wrap(err, "unable to get access groups")
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

func (a *API) DeleteAccessGroup(ctx context.Context, groupID string) error {
	err := a.client.DeleteAccessGroup(ctx, a.CFAccountID, groupID)

	return errors.Wrap(err, "unable to update access groups")
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

	var app *cloudflare.AccessApplication
	for i, g := range apps {
		if g.Domain == domain {
			app = &apps[i]

			break
		}
	}

	return app, nil
}

func (a *API) AccessApplication(ctx context.Context, accessApplicationID string) (cloudflare.AccessApplication, error) {
	cfAG, err := a.client.AccessApplication(ctx, a.CFAccountID, accessApplicationID)

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

func (a *API) DeleteAccessApplication(ctx context.Context, appID string) error {
	err := a.client.DeleteAccessApplication(ctx, a.CFAccountID, appID)

	return errors.Wrap(err, "unable to create access applications")
}

func (a *API) AccessPolicies(ctx context.Context, appID string) (cfcollections.AccessPolicyCollection, error) {
	policies, _, err := a.client.AccessPolicies(ctx, a.CFAccountID, appID, cloudflare.PaginationOptions{})

	policiesCollection := cfcollections.AccessPolicyCollection(policies)

	return policiesCollection, errors.Wrap(err, "unable to get access Policies")
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

func (a *API) ServiceTokens(ctx context.Context) ([]cftypes.ExtendedServiceToken, error) {
	extendedTokens := []cftypes.ExtendedServiceToken{}
	tokens, _, err := a.client.AccessServiceTokens(ctx, a.CFAccountID)
	for _, token := range tokens {
		extendedTokens = append(extendedTokens, cftypes.ExtendedServiceToken{
			AccessServiceToken: cloudflare.AccessServiceToken{
				CreatedAt: token.CreatedAt,
				UpdatedAt: token.UpdatedAt,
				ExpiresAt: token.ExpiresAt,
				ID:        token.ID,
				Name:      token.Name,
				ClientID:  token.ClientID,
			},
		})
	}

	return extendedTokens, errors.Wrap(err, "unable to get service tokens")
}

func (a *API) CreateAccessServiceToken(ctx context.Context, token cftypes.ExtendedServiceToken) (cftypes.ExtendedServiceToken, error) {
	res, err := a.client.CreateAccessServiceToken(ctx, a.CFAccountID, token.Name)
	extendedToken := cftypes.ExtendedServiceToken{
		ClientSecret: res.ClientSecret,
		AccessServiceToken: cloudflare.AccessServiceToken{
			CreatedAt: res.CreatedAt,
			UpdatedAt: res.UpdatedAt,
			ExpiresAt: res.ExpiresAt,
			ID:        res.ID,
			Name:      res.Name,
			ClientID:  res.ClientID,
		},
	}

	return extendedToken, errors.Wrap(err, "unable to create access service token")
}

func (a *API) UpdateAccessServiceToken(ctx context.Context, token cftypes.ExtendedServiceToken) (cftypes.ExtendedServiceToken, error) {
	_, err := a.client.UpdateAccessServiceToken(ctx, a.CFAccountID, token.ID, token.Name)

	return token, errors.Wrap(err, "unable to update access Policy")
}

func (a *API) RotateAccessServiceToken(ctx context.Context, token cftypes.ExtendedServiceToken) (cftypes.ExtendedServiceToken, error) {
	account := cloudflare.AccountIdentifier(a.CFAccountID)
	res, err := a.client.RotateAccessServiceToken(ctx, account, token.ID)

	extendedToken := cftypes.ExtendedServiceToken{
		ClientSecret: res.ClientSecret,
		AccessServiceToken: cloudflare.AccessServiceToken{
			CreatedAt: res.CreatedAt,
			UpdatedAt: res.UpdatedAt,
			ExpiresAt: res.ExpiresAt,
			ID:        res.ID,
			Name:      res.Name,
			ClientID:  res.ClientID,
		},
	}

	return extendedToken, errors.Wrap(err, "unable to update access Policy")
}

func (a *API) DeleteAccessServiceToken(ctx context.Context, tokenID string) error {
	_, err := a.client.DeleteAccessServiceToken(ctx, a.CFAccountID, tokenID)

	return errors.Wrap(err, "unable to update access Policy")
}
