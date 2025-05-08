package cfapi

import (
	"context"

	"github.com/bojanzelic/cloudflare-zero-trust-operator/api/v4alpha1"
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
	//
	iter := a.client.ZeroTrust.Access.Groups.ListAutoPaging(ctx, zero_trust.AccessGroupListParams{
		AccountID: cloudflare.F(a.CFAccountID),
	})

	cfAccessGroupCollection := cfcollections.AccessGroupCollection{}
	for iter.Next() {
		cfAccessGroupCollection = append(cfAccessGroupCollection, iter.Current())
	}

	//
	return cfAccessGroupCollection, errors.Wrap(iter.Err(), "unable to get access groups")
}

func (a *API) AccessGroupByName(ctx context.Context, name string) (*zero_trust.AccessGroupGetResponse, error) {
	//
	iter := a.client.ZeroTrust.Access.Groups.ListAutoPaging(ctx, zero_trust.AccessGroupListParams{
		AccountID: cloudflare.F(a.CFAccountID),
		Name:      cloudflare.F(name),
	})

	//
	iter.Next()
	if iter.Err() != nil {
		empty := zero_trust.AccessGroupGetResponse{}
		return &empty, errors.Wrap(iter.Err(), "unable to get access applications")
	}

	return a.AccessGroup(ctx, iter.Current().ID)
}

func (a *API) AccessGroup(ctx context.Context, accessGroupID string) (*zero_trust.AccessGroupGetResponse, error) {
	//
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
		AccountID: cloudflare.F(a.CFAccountID),
		Name:      cloudflare.F(name),
		Include:   cloudflare.F(include),
		Exclude:   cloudflare.F(exclude),
		Require:   cloudflare.F(require),
	})

	//
	if err != nil {
		dummy := zero_trust.AccessGroupGetResponse{}
		return &dummy, errors.Wrap(err, "unable to create access groups")
	}

	//
	return a.AccessGroup(ctx, insert.ID)
}

func (a *API) UpdateAccessGroup(ctx context.Context,
	groupId string,
	name string,
	include []zero_trust.AccessRuleUnionParam,
	exclude []zero_trust.AccessRuleUnionParam,
	require []zero_trust.AccessRuleUnionParam,
) error {
	//
	_, err := a.client.ZeroTrust.Access.Groups.Update(ctx, groupId, zero_trust.AccessGroupUpdateParams{
		AccountID: cloudflare.F(a.CFAccountID),
		Name:      cloudflare.F(name),
		Include:   cloudflare.F(include),
		Exclude:   cloudflare.F(exclude),
		Require:   cloudflare.F(require),
	})

	return errors.Wrap(err, "unable to update access groups")
}

func (a *API) DeleteAccessGroup(ctx context.Context, groupID string) error {
	//
	_, err := a.client.ZeroTrust.Access.Groups.Delete(ctx, groupID, zero_trust.AccessGroupDeleteParams{
		AccountID: cloudflare.F(a.CFAccountID),
	})

	return errors.Wrap(err, "unable to update access groups")
}

//
//
//

func (a *API) AccessApplications(ctx context.Context) ([]zero_trust.AccessApplicationListResponse, error) {
	//
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
	//
	iter := a.client.ZeroTrust.Access.Applications.ListAutoPaging(ctx, zero_trust.AccessApplicationListParams{
		AccountID: cloudflare.F(a.CFAccountID),
		Domain:    cloudflare.F(domain),
	})

	iter.Next()

	if iter.Err() != nil {
		return nil, errors.Wrap(iter.Err(), "unable to get access applications")
	}

	return a.AccessApplication(ctx, iter.Current().ID)
}

func (a *API) AccessApplication(ctx context.Context, accessApplicationID string) (*zero_trust.AccessApplicationGetResponse, error) {
	//
	cfApp, err := a.client.ZeroTrust.Access.Applications.Get(ctx, accessApplicationID, zero_trust.AccessApplicationGetParams{
		AccountID: cloudflare.F(a.CFAccountID),
	})

	return cfApp, errors.Wrap(err, "unable to get access application")
}

func (a *API) CreateAccessApplication(
	ctx context.Context,
	app *v4alpha1.CloudflareAccessApplication, //nolint:varnamelen
) (*zero_trust.AccessApplicationGetResponse, error) {
	//
	cfApp, err := a.client.ZeroTrust.Access.Applications.New(ctx, zero_trust.AccessApplicationNewParams{
		AccountID: cloudflare.F(a.CFAccountID),
		Body: zero_trust.AccessApplicationNewParamsBody{
			Name:                   cloudflare.F(app.Spec.Name),
			Domain:                 cloudflare.F(app.Spec.Domain),
			Type:                   cloudflare.F(app.Spec.Type),
			AppLauncherVisible:     cloudflare.F(*app.Spec.AppLauncherVisible),
			AllowedIdPs:            cloudflare.F(any(app.Spec.AllowedIdps)),
			AutoRedirectToIdentity: cloudflare.F(*app.Spec.AutoRedirectToIdentity),
			// Policies:			// We rather handle policies definition once created via [ApplyAccessReusablePolicies]
			SessionDuration:         cloudflare.F(app.Spec.SessionDuration),
			EnableBindingCookie:     cloudflare.F(*app.Spec.EnableBindingCookie),
			HTTPOnlyCookieAttribute: cloudflare.F(*app.Spec.HTTPOnlyCookieAttribute),
			LogoURL:                 cloudflare.F(app.Spec.LogoURL),
		},
	})

	if err != nil {
		dummy := zero_trust.AccessApplicationGetResponse{}
		return &dummy, errors.Wrap(err, "unable to create access applications")
	}

	return a.AccessApplication(ctx, cfApp.ID)
}

func (a *API) UpdateAccessApplication(
	ctx context.Context,
	app *v4alpha1.CloudflareAccessApplication, //nolint:varnamelen
) (*zero_trust.AccessApplicationGetResponse, error) {
	//
	cfApp, err := a.client.ZeroTrust.Access.Applications.Update(ctx, app.Status.AccessApplicationID,
		zero_trust.AccessApplicationUpdateParams{
			AccountID: cloudflare.F(a.CFAccountID),
			Body: zero_trust.AccessApplicationUpdateParamsBody{
				Name:                   cloudflare.F(app.Name),
				Domain:                 cloudflare.F(app.Spec.Domain),
				Type:                   cloudflare.F(app.Spec.Type),
				AppLauncherVisible:     cloudflare.F(*app.Spec.AppLauncherVisible),
				AllowedIdPs:            cloudflare.F(any(app.Spec.AllowedIdps)),
				AutoRedirectToIdentity: cloudflare.F(*app.Spec.AutoRedirectToIdentity),
				// Policies:			// We rather handle policies updates in [ApplyAccessReusablePolicies]
				SessionDuration:         cloudflare.F(app.Spec.SessionDuration),
				EnableBindingCookie:     cloudflare.F(*app.Spec.EnableBindingCookie),
				HTTPOnlyCookieAttribute: cloudflare.F(*app.Spec.HTTPOnlyCookieAttribute),
				LogoURL:                 cloudflare.F(app.Spec.LogoURL),
			},
		})

	if err != nil {
		dummy := zero_trust.AccessApplicationGetResponse{}
		return &dummy, errors.Wrap(err, "unable to update access applications")
	}

	return a.AccessApplication(ctx, cfApp.ID)
}

func (a *API) DeleteAccessApplication(ctx context.Context, appID string) error {
	//
	_, err := a.client.ZeroTrust.Access.Applications.Delete(ctx, appID, zero_trust.AccessApplicationDeleteParams{
		AccountID: cloudflare.F(a.CFAccountID),
	})

	return errors.Wrap(err, "unable to create access applications")
}

//
//
//

func (a *API) ApplyAccessReusablePolicies(ctx context.Context, appID string, orderedPolicyIDs []string) (*zero_trust.AccessApplicationUpdateResponse, error) {
	//
	cfApp, err := a.client.ZeroTrust.Access.Applications.Update(ctx, appID, zero_trust.AccessApplicationUpdateParams{
		AccountID: cloudflare.F(a.CFAccountID),
		Body: zero_trust.AccessApplicationUpdateParamsBody{
			Policies: cloudflare.F(any(orderedPolicyIDs)),
		},
	})

	if err != nil {
		dummy := zero_trust.AccessApplicationUpdateResponse{}
		return &dummy, errors.Wrap(err, "unable to update access applications")
	}

	return cfApp, nil
}

//
//
//

func (a *API) AccessReusablePolicy(ctx context.Context, policyID string) (*zero_trust.AccessPolicyGetResponse, error) {
	//
	cfApp, err := a.client.ZeroTrust.Access.Policies.Get(ctx, policyID, zero_trust.AccessPolicyGetParams{
		AccountID: cloudflare.F(a.CFAccountID),
	})

	return cfApp, errors.Wrap(err, "unable to get access reusable policy")
}

func (a *API) CreateAccessReusablePolicy(ctx context.Context, arp *v4alpha1.CloudflareAccessReusablePolicy) (*zero_trust.AccessPolicyGetResponse, error) {
	rp, err := a.client.ZeroTrust.Access.Policies.New(ctx, zero_trust.AccessPolicyNewParams{
		AccountID: cloudflare.F(a.CFAccountID),
		Decision:  cloudflare.F(zero_trust.Decision(arp.Spec.Decision)),
		Name:      cloudflare.F(arp.Spec.Name),
		Include:   cloudflare.F(v4alpha1.ToAccessRuleParams(&arp.Spec.Include)),
		Exclude:   cloudflare.F(v4alpha1.ToAccessRuleParams(&arp.Spec.Exclude)),
		Require:   cloudflare.F(v4alpha1.ToAccessRuleParams(&arp.Spec.Require)),
	})

	if err != nil {
		return nil, errors.Wrap(err, "unable to create access reusable policy")
	}

	return a.AccessReusablePolicy(ctx, rp.ID)
}

func (a *API) UpdateAccessReusablePolicy(ctx context.Context, arp *v4alpha1.CloudflareAccessReusablePolicy) error {
	_, err := a.client.ZeroTrust.Access.Policies.Update(ctx, arp.Status.AccessReusablePolicyID, zero_trust.AccessPolicyUpdateParams{
		AccountID: cloudflare.F(a.CFAccountID),
		Include:   cloudflare.F(v4alpha1.ToAccessRuleParams(&arp.Spec.Include)),
		Exclude:   cloudflare.F(v4alpha1.ToAccessRuleParams(&arp.Spec.Exclude)),
		Require:   cloudflare.F(v4alpha1.ToAccessRuleParams(&arp.Spec.Require)),
	})

	return errors.Wrap(err, "unable to update reusable access policy")
}

func (a *API) DeleteAccessReusablePolicy(ctx context.Context, policyID string) error {

	_, err := a.client.ZeroTrust.Access.Policies.Delete(ctx, policyID, zero_trust.AccessPolicyDeleteParams{
		AccountID: cloudflare.F(a.CFAccountID),
	})

	return errors.Wrap(err, "unable to delete reusable access policy")
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
