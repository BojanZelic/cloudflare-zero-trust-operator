package cfapi

import (
	"context"

	"github.com/bojanzelic/cloudflare-zero-trust-operator/internal/cfcollections"
	"github.com/bojanzelic/cloudflare-zero-trust-operator/internal/cftypes"
	cloudflare "github.com/cloudflare/cloudflare-go/v4"
	"github.com/cloudflare/cloudflare-go/v4/option"
	"github.com/cloudflare/cloudflare-go/v4/zero_trust"
	"github.com/pkg/errors"
)

type API struct {
	CFAccountID string
	client      *cloudflare.Client
}

func New(cfAPIToken string, cfAPIKey string, cfAPIEmail string, cfAccountID string) *API {
	var api *cloudflare.Client

	if cfAPIToken != "" {
		api = cloudflare.NewClient(option.WithAPIToken(cfAPIToken))
	} else {
		api = cloudflare.NewClient(option.WithAPIKey(cfAPIKey), option.WithAPIEmail(cfAPIEmail))
	}

	return &API{
		CFAccountID: cfAccountID,
		client:      api,
	}
}

func (a *API) AccessGroups(ctx context.Context) (cfcollections.AccessGroupCollection, error) {
	cfAccessGroups, err := a.client.ZeroTrust.Access.Groups.List(ctx, zero_trust.AccessGroupListParams{
		AccountID: cloudflare.F(a.CFAccountID),
	})
	cfAccessGroupCollection := cfcollections.AccessGroupCollection(cfAccessGroups)

	return cfAccessGroupCollection, errors.Wrap(err, "unable to get access groups")
}

func (a *API) AccessGroup(ctx context.Context, accessGroupID string) (cloudflare.AccessGroup, error) {

	cfAG, err := a.client.ZeroTrust.Access.Groups.Get(ctx, accessGroupID, zero_trust.AccessGroupGetParams{
		AccountID: cloudflare.F(a.CFAccountID),
	})

	return cfAG, errors.Wrap(err, "unable to get access group")
}

func (a *API) CreateAccessGroup(ctx context.Context, ag cloudflare.AccessGroup) (cloudflare.AccessGroup, error) {

	params := cloudflare.CreateAccessGroupParams{
		Name:    ag.Name,
		Include: ag.Include,
		Exclude: ag.Exclude,
		Require: ag.Require,
	}

	cfAG, err := a.client.ZeroTrust.Access.Groups.New(ctx, zero_trust.AccessGroupNewParams{
		AccountID: cloudflare.F(a.CFAccountID),
	})

	return cfAG, errors.Wrap(err, "unable to create access groups")
}

func (a *API) UpdateAccessGroup(ctx context.Context, ag cloudflare.AccessGroup) (cloudflare.AccessGroup, error) {

	params := cloudflare.UpdateAccessGroupParams{
		ID:      ag.ID,
		Name:    ag.Name,
		Include: ag.Include,
		Exclude: ag.Exclude,
		Require: ag.Require,
	}

	cfAG, err := a.client.ZeroTrust.Access.Groups.Update(ctx, ag.ID, zero_trust.AccessGroupUpdateParams{
		AccountID: cloudflare.F(a.CFAccountID),
	})

	return cfAG, errors.Wrap(err, "unable to update access groups")
}

func (a *API) DeleteAccessGroup(ctx context.Context, groupID string) error {

	_, err := a.client.ZeroTrust.Access.Groups.Delete(ctx, groupID, zero_trust.AccessGroupDeleteParams{
		AccountID: cloudflare.F(a.CFAccountID),
	})

	return errors.Wrap(err, "unable to update access groups")
}

func (a *API) AccessApplications(ctx context.Context) ([]cloudflare.AccessApplication, error) {

	apps, err := a.client.ZeroTrust.Access.Applications.List(ctx, zero_trust.AccessApplicationListParams{
		AccountID: cloudflare.F(a.CFAccountID),
	})

	return apps, errors.Wrap(err, "unable to get access applications")
}

func (a *API) FindAccessApplicationByDomain(ctx context.Context, domain string) (*cloudflare.AccessApplication, error) {

	apps, err := a.client.ZeroTrust.Access.Applications.List(ctx, zero_trust.AccessApplicationListParams{
		AccountID: cloudflare.F(a.CFAccountID),
	})
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

	cfAG, err := a.client.ZeroTrust.Access.Applications.Get(ctx, accessApplicationID, zero_trust.AccessApplicationGetParams{
		AccountID: cloudflare.F(a.CFAccountID),
	})

	return cfAG, errors.Wrap(err, "unable to get access application")
}

func (a *API) CreateAccessApplication(ctx context.Context, ag cloudflare.AccessApplication) (cloudflare.AccessApplication, error) {

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
		Destinations:                   ag.Destinations,
		ServiceAuth401Redirect:         ag.ServiceAuth401Redirect,
		SessionDuration:                ag.SessionDuration,
		SkipInterstitial:               ag.SkipInterstitial,
		Type:                           ag.Type,
		CustomPages:                    ag.CustomPages,
		Tags:                           ag.Tags,
		AccessAppLauncherCustomization: ag.AccessAppLauncherCustomization,
	}

	cfAG, err := a.client.ZeroTrust.Access.Applications.New(ctx, zero_trust.AccessApplicationNewParams{
		AccountID: cloudflare.F(a.CFAccountID),
	})

	return cfAG, errors.Wrap(err, "unable to create access applications")
}

func (a *API) UpdateAccessApplication(ctx context.Context, ag cloudflare.AccessApplication) (cloudflare.AccessApplication, error) {

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
		Destinations:                   ag.Destinations,
		ServiceAuth401Redirect:         ag.ServiceAuth401Redirect,
		SessionDuration:                ag.SessionDuration,
		SkipInterstitial:               ag.SkipInterstitial,
		Type:                           ag.Type,
		CustomPages:                    ag.CustomPages,
		Tags:                           ag.Tags,
		AccessAppLauncherCustomization: ag.AccessAppLauncherCustomization,
	}
	cfAG, err := a.client.ZeroTrust.Access.Applications.Update(ctx, ag.ID, zero_trust.AccessApplicationUpdateParams{
		AccountID: cloudflare.F(a.CFAccountID),
	})

	return cfAG, errors.Wrap(err, "unable to update access applications")
}

func (a *API) DeleteAccessApplication(ctx context.Context, appID string) error {

	_, err := a.client.ZeroTrust.Access.Applications.Delete(ctx, appID, zero_trust.AccessApplicationDeleteParams{
		AccountID: cloudflare.F(a.CFAccountID),
	})

	return errors.Wrap(err, "unable to create access applications")
}

func (a *API) AccessPolicies(ctx context.Context, appID string) (cfcollections.AccessPolicyCollection, error) {

	policies, err := a.client.ZeroTrust.Access.Policies.List(ctx, zero_trust.AccessPolicyListParams{
		AccountID: cloudflare.F(a.CFAccountID),
	})

	policiesCollection := cfcollections.AccessPolicyCollection(policies)

	return policiesCollection, errors.Wrap(err, "unable to get access Policies")
}

func (a *API) CreateAccessPolicy(ctx context.Context, appID string, ag cloudflare.AccessPolicy) (cloudflare.AccessPolicy, error) {

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
	cfAG, err := a.client.ZeroTrust.Access.Policies.New(ctx, zero_trust.AccessPolicyNewParams{
		AccountID: cloudflare.F(a.CFAccountID),
	})

	return cfAG, errors.Wrap(err, "unable to create access Policy")
}

func (a *API) UpdateAccessPolicy(ctx context.Context, appID string, ag cloudflare.AccessPolicy) (cloudflare.AccessPolicy, error) {

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
	cfAG, err := a.client.ZeroTrust.Access.Policies.Update(ctx, ag.ID, zero_trust.AccessPolicyUpdateParams{
		AccountID: cloudflare.F(a.CFAccountID),
	})

	return cfAG, errors.Wrap(err, "unable to update access Policy")
}

func (a *API) DeleteAccessPolicy(ctx context.Context, appID string, policyID string) error {

	params := cloudflare.DeleteAccessPolicyParams{
		ApplicationID: appID,
		PolicyID:      policyID,
	}
	_, err := a.client.ZeroTrust.Access.Policies.Delete(ctx, policyID, zero_trust.AccessPolicyDeleteParams{
		AccountID: cloudflare.F(a.CFAccountID),
	})

	return errors.Wrap(err, "unable to update access Policy")
}

func (a *API) ServiceTokens(ctx context.Context) ([]cftypes.ExtendedServiceToken, error) {

	extendedTokens := []cftypes.ExtendedServiceToken{}
	tokens, err := a.client.ZeroTrust.Access.ServiceTokens.List(ctx, zero_trust.AccessServiceTokenListParams{
		AccountID: cloudflare.F(a.CFAccountID),
	})
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

	params := cloudflare.CreateAccessServiceTokenParams{
		Name: token.Name,
	}

	res, err := a.client.ZeroTrust.Access.ServiceTokens.New(ctx, zero_trust.AccessServiceTokenNewParams{
		AccountID: cloudflare.F(a.CFAccountID),
	})
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

	params := cloudflare.UpdateAccessServiceTokenParams{
		Name: token.Name,
		UUID: token.ID,
	}

	_, err := a.client.ZeroTrust.Access.ServiceTokens.New(ctx, zero_trust.AccessServiceTokenNewParams{
		AccountID: cloudflare.F(a.CFAccountID),
	})

	return token, errors.Wrap(err, "unable to update access Policy")
}

func (a *API) RotateAccessServiceToken(ctx context.Context, token cftypes.ExtendedServiceToken) (cftypes.ExtendedServiceToken, error) {
	res, err := a.client.ZeroTrust.Access.ServiceTokens.Rotate(ctx, token.ID, zero_trust.AccessServiceTokenRotateParams{
		AccountID: cloudflare.F(a.CFAccountID),
	})

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

	_, err := a.client.ZeroTrust.Access.ServiceTokens.Delete(ctx, tokenID, zero_trust.AccessServiceTokenDeleteParams{
		AccountID: cloudflare.F(a.CFAccountID),
	})

	return errors.Wrap(err, "unable to update access Policy")
}
