package cfapi

import (
	"context"

	cloudflare "github.com/cloudflare/cloudflare-go"
	"github.com/kadaan/cloudflare-zero-trust-operator/internal/cfcollections"
	"github.com/kadaan/cloudflare-zero-trust-operator/internal/cftypes"
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
	account := cloudflare.AccountIdentifier(a.CFAccountID)
	cfAccessGroups, _, err := a.client.ListAccessGroups(ctx, account, cloudflare.ListAccessGroupsParams{})
	cfAccessGroupCollection := cfcollections.AccessGroupCollection(cfAccessGroups)

	return cfAccessGroupCollection, errors.Wrap(err, "unable to get access groups")
}

func (a *API) AccessGroup(ctx context.Context, accessGroupID string) (cloudflare.AccessGroup, error) {
	account := cloudflare.AccountIdentifier(a.CFAccountID)

	cfAG, err := a.client.GetAccessGroup(ctx, account, accessGroupID)

	return cfAG, errors.Wrap(err, "unable to get access group")
}

func (a *API) CreateAccessGroup(ctx context.Context, ag cloudflare.AccessGroup) (cloudflare.AccessGroup, error) {
	account := cloudflare.AccountIdentifier(a.CFAccountID)

	params := cloudflare.CreateAccessGroupParams{
		Name:    ag.Name,
		Include: ag.Include,
		Exclude: ag.Exclude,
		Require: ag.Require,
	}

	cfAG, err := a.client.CreateAccessGroup(ctx, account, params)

	return cfAG, errors.Wrap(err, "unable to create access groups")
}

func (a *API) UpdateAccessGroup(ctx context.Context, ag cloudflare.AccessGroup) (cloudflare.AccessGroup, error) {
	account := cloudflare.AccountIdentifier(a.CFAccountID)

	params := cloudflare.UpdateAccessGroupParams{
		ID:      ag.ID,
		Name:    ag.Name,
		Include: ag.Include,
		Exclude: ag.Exclude,
		Require: ag.Require,
	}

	cfAG, err := a.client.UpdateAccessGroup(ctx, account, params)

	return cfAG, errors.Wrap(err, "unable to update access groups")
}

func (a *API) DeleteAccessGroup(ctx context.Context, groupID string) error {
	account := cloudflare.AccountIdentifier(a.CFAccountID)

	err := a.client.DeleteAccessGroup(ctx, account, groupID)

	return errors.Wrap(err, "unable to update access groups")
}

func (a *API) AccessApplications(ctx context.Context) ([]cloudflare.AccessApplication, error) {
	account := cloudflare.AccountIdentifier(a.CFAccountID)

	apps, _, err := a.client.ListAccessApplications(ctx, account, cloudflare.ListAccessApplicationsParams{})

	return apps, errors.Wrap(err, "unable to get access applications")
}

func (a *API) FindAccessApplicationByDomain(ctx context.Context, domain string) (*cloudflare.AccessApplication, error) {
	account := cloudflare.AccountIdentifier(a.CFAccountID)

	apps, _, err := a.client.ListAccessApplications(ctx, account, cloudflare.ListAccessApplicationsParams{})
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
	account := cloudflare.AccountIdentifier(a.CFAccountID)

	cfAG, err := a.client.GetAccessApplication(ctx, account, accessApplicationID)

	return cfAG, errors.Wrap(err, "unable to get access application")
}

func (a *API) CreateAccessApplication(ctx context.Context, ag cloudflare.AccessApplication) (cloudflare.AccessApplication, error) {
	account := cloudflare.AccountIdentifier(a.CFAccountID)

	params := cloudflare.CreateAccessApplicationParams{
		AllowedIdps:                    ag.AllowedIdps,
		AppLauncherVisible:             ag.AppLauncherVisible,
		AUD:                            ag.AUD,
		AutoRedirectToIdentity:         ag.AutoRedirectToIdentity,
		CorsHeaders:                    ag.CorsHeaders,
		CustomDenyMessage:              ag.CustomDenyMessage,
		CustomDenyURL:                  ag.CustomDenyURL,
		CustomNonIdentityDenyURL:       ag.CustomNonIdentityDenyURL,
		Domain:                         ag.Domain,
		EnableBindingCookie:            ag.EnableBindingCookie,
		GatewayRules:                   ag.GatewayRules,
		HttpOnlyCookieAttribute:        ag.HttpOnlyCookieAttribute,
		LogoURL:                        ag.LogoURL,
		Name:                           ag.Name,
		PathCookieAttribute:            ag.PathCookieAttribute,
		PrivateAddress:                 ag.PrivateAddress,
		SaasApplication:                ag.SaasApplication,
		SameSiteCookieAttribute:        ag.SameSiteCookieAttribute,
		SelfHostedDomains:              ag.SelfHostedDomains,
		ServiceAuth401Redirect:         ag.ServiceAuth401Redirect,
		SessionDuration:                ag.SessionDuration,
		SkipInterstitial:               ag.SkipInterstitial,
		Type:                           ag.Type,
		CustomPages:                    ag.CustomPages,
		Tags:                           ag.Tags,
		AccessAppLauncherCustomization: ag.AccessAppLauncherCustomization,
	}

	cfAG, err := a.client.CreateAccessApplication(ctx, account, params)

	return cfAG, errors.Wrap(err, "unable to create access applications")
}

func (a *API) UpdateAccessApplication(ctx context.Context, ag cloudflare.AccessApplication) (cloudflare.AccessApplication, error) {
	account := cloudflare.AccountIdentifier(a.CFAccountID)

	params := cloudflare.UpdateAccessApplicationParams{
		ID:                             ag.ID,
		AllowedIdps:                    ag.AllowedIdps,
		AppLauncherVisible:             ag.AppLauncherVisible,
		AUD:                            ag.AUD,
		AutoRedirectToIdentity:         ag.AutoRedirectToIdentity,
		CorsHeaders:                    ag.CorsHeaders,
		CustomDenyMessage:              ag.CustomDenyMessage,
		CustomDenyURL:                  ag.CustomDenyURL,
		CustomNonIdentityDenyURL:       ag.CustomNonIdentityDenyURL,
		Domain:                         ag.Domain,
		EnableBindingCookie:            ag.EnableBindingCookie,
		GatewayRules:                   ag.GatewayRules,
		HttpOnlyCookieAttribute:        ag.HttpOnlyCookieAttribute,
		LogoURL:                        ag.LogoURL,
		Name:                           ag.Name,
		PathCookieAttribute:            ag.PathCookieAttribute,
		PrivateAddress:                 ag.PrivateAddress,
		SaasApplication:                ag.SaasApplication,
		SameSiteCookieAttribute:        ag.SameSiteCookieAttribute,
		SelfHostedDomains:              ag.SelfHostedDomains,
		ServiceAuth401Redirect:         ag.ServiceAuth401Redirect,
		SessionDuration:                ag.SessionDuration,
		SkipInterstitial:               ag.SkipInterstitial,
		Type:                           ag.Type,
		CustomPages:                    ag.CustomPages,
		Tags:                           ag.Tags,
		AccessAppLauncherCustomization: ag.AccessAppLauncherCustomization,
	}
	cfAG, err := a.client.UpdateAccessApplication(ctx, account, params)

	return cfAG, errors.Wrap(err, "unable to update access applications")
}

func (a *API) DeleteAccessApplication(ctx context.Context, appID string) error {
	account := cloudflare.AccountIdentifier(a.CFAccountID)

	err := a.client.DeleteAccessApplication(ctx, account, appID)

	return errors.Wrap(err, "unable to create access applications")
}

func (a *API) AccessPolicies(ctx context.Context, appID string) (cfcollections.AccessPolicyCollection, error) {
	account := cloudflare.AccountIdentifier(a.CFAccountID)

	policies, _, err := a.client.ListAccessPolicies(ctx, account, cloudflare.ListAccessPoliciesParams{ApplicationID: appID})

	policiesCollection := cfcollections.AccessPolicyCollection(policies)

	return policiesCollection, errors.Wrap(err, "unable to get access Policies")
}

func (a *API) CreateAccessPolicy(ctx context.Context, appID string, ag cloudflare.AccessPolicy) (cloudflare.AccessPolicy, error) {
	account := cloudflare.AccountIdentifier(a.CFAccountID)

	params := cloudflare.CreateAccessPolicyParams{
		ApplicationID:                appID,
		Precedence:                   ag.Precedence,
		Decision:                     ag.Decision,
		Name:                         ag.Name,
		IsolationRequired:            ag.IsolationRequired,
		SessionDuration:              ag.SessionDuration,
		PurposeJustificationRequired: ag.PurposeJustificationRequired,
		PurposeJustificationPrompt:   ag.PurposeJustificationPrompt,
		ApprovalRequired:             ag.ApprovalRequired,
		ApprovalGroups:               ag.ApprovalGroups,
		Include:                      ag.Include,
		Exclude:                      ag.Exclude,
		Require:                      ag.Require,
	}
	cfAG, err := a.client.CreateAccessPolicy(ctx, account, params)

	return cfAG, errors.Wrap(err, "unable to create access Policy")
}

func (a *API) UpdateAccessPolicy(ctx context.Context, appID string, ag cloudflare.AccessPolicy) (cloudflare.AccessPolicy, error) {
	account := cloudflare.AccountIdentifier(a.CFAccountID)

	params := cloudflare.UpdateAccessPolicyParams{
		ApplicationID:                appID,
		PolicyID:                     ag.ID,
		Precedence:                   ag.Precedence,
		Decision:                     ag.Decision,
		Name:                         ag.Name,
		IsolationRequired:            ag.IsolationRequired,
		SessionDuration:              ag.SessionDuration,
		PurposeJustificationRequired: ag.PurposeJustificationRequired,
		PurposeJustificationPrompt:   ag.PurposeJustificationPrompt,
		ApprovalRequired:             ag.ApprovalRequired,
		ApprovalGroups:               ag.ApprovalGroups,
		Include:                      ag.Include,
		Exclude:                      ag.Exclude,
		Require:                      ag.Require,
	}
	cfAG, err := a.client.UpdateAccessPolicy(ctx, account, params)

	return cfAG, errors.Wrap(err, "unable to update access Policy")
}

func (a *API) DeleteAccessPolicy(ctx context.Context, appID string, policyID string) error {
	account := cloudflare.AccountIdentifier(a.CFAccountID)

	params := cloudflare.DeleteAccessPolicyParams{
		ApplicationID: appID,
		PolicyID:      policyID,
	}
	err := a.client.DeleteAccessPolicy(ctx, account, params)

	return errors.Wrap(err, "unable to update access Policy")
}

func (a *API) ServiceTokens(ctx context.Context) ([]cftypes.ExtendedServiceToken, error) {
	account := cloudflare.AccountIdentifier(a.CFAccountID)

	extendedTokens := []cftypes.ExtendedServiceToken{}
	tokens, _, err := a.client.ListAccessServiceTokens(ctx, account, cloudflare.ListAccessServiceTokensParams{})
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
	account := cloudflare.AccountIdentifier(a.CFAccountID)

	params := cloudflare.CreateAccessServiceTokenParams{
		Name: token.Name,
	}

	res, err := a.client.CreateAccessServiceToken(ctx, account, params)
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
	account := cloudflare.AccountIdentifier(a.CFAccountID)

	params := cloudflare.UpdateAccessServiceTokenParams{
		Name: token.Name,
		UUID: token.ID,
	}

	_, err := a.client.UpdateAccessServiceToken(ctx, account, params)

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
	account := cloudflare.AccountIdentifier(a.CFAccountID)

	_, err := a.client.DeleteAccessServiceToken(ctx, account, tokenID)

	return errors.Wrap(err, "unable to update access Policy")
}
