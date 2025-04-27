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

//
//
//

func (a *API) AccessGroups(ctx context.Context) (cfcollections.AccessGroupCollection, error) {
	iter := a.client.ZeroTrust.Access.Groups.ListAutoPaging(ctx, zero_trust.AccessGroupListParams{
		AccountID: cloudflare.F(a.CFAccountID),
	})

	cfAccessGroupCollection := cfcollections.AccessGroupCollection{}
	for iter.Next() {
		cfAccessGroupCollection = append(cfAccessGroupCollection, iter.Current())
	}

	return cfAccessGroupCollection, errors.Wrap(iter.Err(), "unable to get access groups")
}

func (a *API) AccessGroupByName(ctx context.Context, name string) (*zero_trust.AccessGroupGetResponse, error) {

	iter := a.client.ZeroTrust.Access.Groups.ListAutoPaging(ctx, zero_trust.AccessGroupListParams{
		AccountID: cloudflare.F(a.CFAccountID),
		Name:      cloudflare.F(name),
	})

	iter.Next()

	if iter.Err() != nil {
		empty := zero_trust.AccessGroupGetResponse{}
		return &empty, errors.Wrap(iter.Err(), "unable to get access applications")
	}

	return a.AccessGroup(ctx, iter.Current().ID)
}

func (a *API) AccessGroup(ctx context.Context, accessGroupID string) (*zero_trust.AccessGroupGetResponse, error) {

	cfAG, err := a.client.ZeroTrust.Access.Groups.Get(ctx, accessGroupID, zero_trust.AccessGroupGetParams{
		AccountID: cloudflare.F(a.CFAccountID),
	})

	return cfAG, errors.Wrap(err, "unable to get access group")
}

func (a *API) CreateAccessGroup(ctx context.Context,
	name string,
	include []zero_trust.AccessRuleUnionParam,
	exclude []zero_trust.AccessRuleUnionParam,
	require []zero_trust.AccessRuleUnionParam,
) (*zero_trust.AccessGroupGetResponse, error) {
	//
	insert, err := a.client.ZeroTrust.Access.Groups.New(ctx, zero_trust.AccessGroupNewParams{
		Name:      cloudflare.F(name),
		AccountID: cloudflare.F(a.CFAccountID),
		Include:   cloudflare.F(include),
		Exclude:   cloudflare.F(exclude),
		Require:   cloudflare.F(require),
	})
	if err != nil {
		dummy := zero_trust.AccessGroupGetResponse{}
		return &dummy, errors.Wrap(err, "unable to create access groups")
	}

	return a.AccessGroup(ctx, insert.ID)
}

func (a *API) UpdateAccessGroup(ctx context.Context, ag cloudflare.AccessGroup) (*zero_trust.AccessGroupGetResponse, error) {

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

	if err != nil {
		dummy := zero_trust.AccessGroupGetResponse{}
		return &dummy, errors.Wrap(err, "unable to update access groups")
	}

	return a.AccessGroup(ctx, cfAG.ID)
}

func (a *API) DeleteAccessGroup(ctx context.Context, groupID string) error {

	_, err := a.client.ZeroTrust.Access.Groups.Delete(ctx, groupID, zero_trust.AccessGroupDeleteParams{
		AccountID: cloudflare.F(a.CFAccountID),
	})

	return errors.Wrap(err, "unable to update access groups")
}

//
//
//

func (a *API) AccessApplications(ctx context.Context) ([]zero_trust.AccessApplicationListResponse, error) {

	iter := a.client.ZeroTrust.Access.Applications.ListAutoPaging(ctx, zero_trust.AccessApplicationListParams{
		AccountID: cloudflare.F(a.CFAccountID),
	})

	apps := []zero_trust.AccessApplicationListResponse{}
	for iter.Next() {
		apps = append(apps, iter.Current())
	}

	return apps, errors.Wrap(iter.Err(), "unable to get access applications")
}

func (a *API) FindAccessApplicationByDomain(ctx context.Context, domain string) (*zero_trust.AccessApplicationGetResponse, error) {
	iter := a.client.ZeroTrust.Access.Applications.ListAutoPaging(ctx, zero_trust.AccessApplicationListParams{
		AccountID: cloudflare.F(a.CFAccountID),
		Domain:    cloudflare.F(domain),
	})

	iter.Next()

	if iter.Err() != nil {
		empty := zero_trust.AccessApplicationGetResponse{}
		return &empty, errors.Wrap(iter.Err(), "unable to get access applications")
	}

	return a.AccessApplication(ctx, iter.Current().ID)
}

func (a *API) AccessApplication(ctx context.Context, accessApplicationID string) (*zero_trust.AccessApplicationGetResponse, error) {

	cfAG, err := a.client.ZeroTrust.Access.Applications.Get(ctx, accessApplicationID, zero_trust.AccessApplicationGetParams{
		AccountID: cloudflare.F(a.CFAccountID),
	})

	return cfAG, errors.Wrap(err, "unable to get access application")
}

func (a *API) CreateAccessApplication(ctx context.Context, ag cloudflare.AccessApplication) (*zero_trust.AccessApplicationGetResponse, error) {

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

	if err != nil {
		dummy := zero_trust.AccessApplicationGetResponse{}
		return &dummy, errors.Wrap(err, "unable to create access applications")
	}

	return a.AccessApplication(ctx, cfAG.ID)
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

//
//
//

func (a *API) LegacyAccessPolicies(ctx context.Context, appID string) (cfcollections.LegacyAccessPolicyCollection, error) {

	iter := a.client.ZeroTrust.Access.Applications.Policies.ListAutoPaging(ctx, appID, zero_trust.AccessApplicationPolicyListParams{
		AccountID: cloudflare.F(a.CFAccountID),
	})

	policiesCollection := cfcollections.LegacyAccessPolicyCollection{}
	for iter.Next() {
		policiesCollection = append(policiesCollection, iter.Current())
	}

	return policiesCollection, errors.Wrap(iter.Err(), "unable to get access Policies")
}

func (a *API) CreateLegacyAccessPolicies(ctx context.Context, appID string, lap cloudflare.AccessPolicy) (cloudflare.AccessPolicy, error) {

	params := cloudflare.CreateAccessPolicyParams{
		ApplicationID:                appID,
		Precedence:                   lap.Precedence,
		Decision:                     lap.Decision,
		Name:                         lap.Name,
		IsolationRequired:            lap.IsolationRequired,
		SessionDuration:              lap.SessionDuration,
		PurposeJustificationRequired: lap.PurposeJustificationRequired,
		PurposeJustificationPrompt:   lap.PurposeJustificationPrompt,
		ApprovalRequired:             lap.ApprovalRequired,
		ApprovalGroups:               lap.ApprovalGroups,
		Include:                      lap.Include,
		Exclude:                      lap.Exclude,
		Require:                      lap.Require,
	}
	cfAG, err := a.client.ZeroTrust.Access.Applications.Policies.New(ctx, appID, zero_trust.AccessApplicationPolicyNewParams{
		AccountID: cloudflare.F(a.CFAccountID),
	})

	return cfAG, errors.Wrap(err, "unable to create access Policy")
}

func (a *API) UpdateLegacyAccessPolicy(ctx context.Context, appID string, lap zero_trust.AccessApplicationPolicyListResponse) (*zero_trust.AccessApplicationPolicyGetResponse, error) {

	params := cloudflare.UpdateAccessPolicyParams{
		ApplicationID:                appID,
		PolicyID:                     lap.ID,
		Precedence:                   lap.Precedence,
		Decision:                     lap.Decision,
		Name:                         lap.Name,
		IsolationRequired:            lap.IsolationRequired,
		SessionDuration:              lap.SessionDuration,
		PurposeJustificationRequired: lap.PurposeJustificationRequired,
		PurposeJustificationPrompt:   lap.PurposeJustificationPrompt,
		ApprovalRequired:             lap.ApprovalRequired,
		ApprovalGroups:               lap.ApprovalGroups,
		Include:                      lap.Include,
		Exclude:                      lap.Exclude,
		Require:                      lap.Require,
	}
	update, err := a.client.ZeroTrust.Access.Applications.Policies.Update(ctx, appID, lap.ID, zero_trust.AccessApplicationPolicyUpdateParams{
		AccountID: cloudflare.F(a.CFAccountID),
	})

	if err != nil {
		dummy := zero_trust.AccessApplicationPolicyGetResponse{}
		return &dummy, errors.Wrap(err, "unable to update access Policy")
	}

	get, err := a.client.ZeroTrust.Access.Applications.Policies.Get(ctx, appID, update.ID, zero_trust.AccessApplicationPolicyGetParams{
		AccountID: cloudflare.F(a.CFAccountID),
	})

	return get, errors.Wrap(err, "unable to update access Policy")
}

func (a *API) DeleteLegacyAccessPolicy(ctx context.Context, appID string, policyID string) error {

	_, err := a.client.ZeroTrust.Access.Applications.Policies.Delete(ctx, appID, policyID, zero_trust.AccessApplicationPolicyDeleteParams{
		AccountID: cloudflare.F(a.CFAccountID),
	})

	return errors.Wrap(err, "unable to update access Policy")
}

//
//
//

func (a *API) ServiceTokens(ctx context.Context) ([]cftypes.ExtendedServiceToken, error) {

	iter := a.client.ZeroTrust.Access.ServiceTokens.ListAutoPaging(ctx, zero_trust.AccessServiceTokenListParams{
		AccountID: cloudflare.F(a.CFAccountID),
	})

	extendedTokens := []cftypes.ExtendedServiceToken{}
	for iter.Next() {
		extendedTokens = append(extendedTokens, cftypes.ExtendedServiceToken{
			ServiceToken: iter.Current(),
		})
	}

	return extendedTokens, errors.Wrap(iter.Err(), "unable to get service tokens")
}

func (a *API) CreateAccessServiceToken(ctx context.Context, token cftypes.ExtendedServiceToken) (cftypes.ExtendedServiceToken, error) {
	//
	res, err := a.client.ZeroTrust.Access.ServiceTokens.New(ctx, zero_trust.AccessServiceTokenNewParams{
		AccountID: cloudflare.F(a.CFAccountID),
		Name:      cloudflare.F(token.Name),
	})

	if err != nil {
		return cftypes.ExtendedServiceToken{}, errors.Wrap(err, "unable to create access service token")
	}

	sToken, err := a.client.ZeroTrust.Access.ServiceTokens.Get(ctx, res.ID, zero_trust.AccessServiceTokenGetParams{
		AccountID: cloudflare.F(a.CFAccountID),
	})

	extendedToken := cftypes.ExtendedServiceToken{
		ClientSecret: res.ClientSecret,
		ServiceToken: zero_trust.ServiceToken{
			CreatedAt: sToken.CreatedAt,
			UpdatedAt: sToken.UpdatedAt,
			ExpiresAt: sToken.ExpiresAt,
			ID:        sToken.ID,
			Name:      sToken.Name,
			ClientID:  sToken.ClientID,
		},
	}

	return extendedToken, errors.Wrap(err, "unable to create access service token")
}

func (a *API) DeleteAccessServiceToken(ctx context.Context, tokenID string) error {

	_, err := a.client.ZeroTrust.Access.ServiceTokens.Delete(ctx, tokenID, zero_trust.AccessServiceTokenDeleteParams{
		AccountID: cloudflare.F(a.CFAccountID),
	})

	return errors.Wrap(err, "unable to update access Policy")
}
