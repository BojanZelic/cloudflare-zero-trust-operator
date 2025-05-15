package cfapi

import (
	"context"

	"github.com/bojanzelic/cloudflare-zero-trust-operator/api/v4alpha1"
	"github.com/bojanzelic/cloudflare-zero-trust-operator/internal/cftypes"
	cloudflare "github.com/cloudflare/cloudflare-go/v4"
	"github.com/cloudflare/cloudflare-go/v4/option"
	"github.com/cloudflare/cloudflare-go/v4/zero_trust"
	"github.com/pkg/errors"
)

type API struct {
	CFAccountID    string
	client         *cloudflare.Client
	optionalTracer *InsertedCFRessourcesTracer
}

func New(cfAPIToken string, cfAPIKey string, cfAPIEmail string, cfAccountID string, optionalTracer *InsertedCFRessourcesTracer) *API {
	var api *cloudflare.Client

	if cfAPIToken != "" {
		api = cloudflare.NewClient(option.WithAPIToken(cfAPIToken))
	} else {
		api = cloudflare.NewClient(option.WithAPIKey(cfAPIKey), option.WithAPIEmail(cfAPIEmail))
	}

	return &API{
		CFAccountID:    cfAccountID,
		client:         api,
		optionalTracer: optionalTracer,
	}
}

//
//
//

func (a *API) AccessGroupByName(ctx context.Context, name string) (*zero_trust.AccessGroupGetResponse, error) {
	//
	iter := a.client.ZeroTrust.Access.Groups.ListAutoPaging(ctx, zero_trust.AccessGroupListParams{
		AccountID: cloudflare.F(a.CFAccountID),
		Name:      cloudflare.F(name),
	})

	for iter.Next() {
		// return first
		return a.AccessGroup(ctx, iter.Current().ID)
	}

	//
	return nil, errors.Wrap(iter.Err(), "unable to get access group by name")
}

func (a *API) AccessGroup(ctx context.Context, accessGroupID string) (*zero_trust.AccessGroupGetResponse, error) {
	//
	cfAG, err := a.client.ZeroTrust.Access.Groups.Get(ctx, accessGroupID, zero_trust.AccessGroupGetParams{
		AccountID: cloudflare.F(a.CFAccountID),
	})

	return cfAG, errors.Wrap(err, "unable to get access group")
}

func (a *API) CreateAccessGroup(ctx context.Context, group *v4alpha1.CloudflareAccessGroup) (*zero_trust.AccessGroupGetResponse, error) {
	//
	insert, err := a.client.ZeroTrust.Access.Groups.New(ctx,
		zero_trust.AccessGroupNewParams{
			AccountID: cloudflare.F(a.CFAccountID),
			Name:      cloudflare.F(group.Spec.Name),
			Include:   cloudflare.F(group.Spec.Include.ToAccessRuleParams(group.Status.ResolvedIdpsFromRefs.Include)),
			Exclude:   cloudflare.F(group.Spec.Exclude.ToAccessRuleParams(group.Status.ResolvedIdpsFromRefs.Exclude)),
			Require:   cloudflare.F(group.Spec.Require.ToAccessRuleParams(group.Status.ResolvedIdpsFromRefs.Require)),
		},
	)

	//
	if err != nil {
		return nil, errors.Wrap(err, "unable to create access group")
	}

	//
	if a.optionalTracer != nil {
		a.optionalTracer.GroupInserted(insert.ID)
	}

	//
	return a.AccessGroup(ctx, insert.ID)
}

func (a *API) UpdateAccessGroup(ctx context.Context, group *v4alpha1.CloudflareAccessGroup) error {
	//
	_, err := a.client.ZeroTrust.Access.Groups.Update(ctx, group.Status.AccessGroupID,
		zero_trust.AccessGroupUpdateParams{
			AccountID: cloudflare.F(a.CFAccountID),
			Name:      cloudflare.F(group.Spec.Name),
			Include:   cloudflare.F(group.Spec.Include.ToAccessRuleParams(group.Status.ResolvedIdpsFromRefs.Include)),
			Exclude:   cloudflare.F(group.Spec.Exclude.ToAccessRuleParams(group.Status.ResolvedIdpsFromRefs.Exclude)),
			Require:   cloudflare.F(group.Spec.Require.ToAccessRuleParams(group.Status.ResolvedIdpsFromRefs.Require)),
		},
	)

	return errors.Wrap(err, "unable to update access group")
}

func (a *API) DeleteAccessGroup(ctx context.Context, groupID string) error {
	//
	_, err := a.client.ZeroTrust.Access.Groups.Delete(ctx, groupID, zero_trust.AccessGroupDeleteParams{
		AccountID: cloudflare.F(a.CFAccountID),
	})

	//
	if a.optionalTracer != nil && err == nil {
		a.optionalTracer.GroupDeleted(groupID)
	}

	return errors.Wrap(err, "unable to delete access group")
}

//
//
//

func (a *API) FindAccessApplicationByDomain(ctx context.Context, domain string) (*zero_trust.AccessApplicationGetResponse, error) {
	//
	iter := a.client.ZeroTrust.Access.Applications.ListAutoPaging(ctx, zero_trust.AccessApplicationListParams{
		AccountID: cloudflare.F(a.CFAccountID),
		Domain:    cloudflare.F(domain),
	})

	for iter.Next() {
		// return first
		return a.AccessApplication(ctx, iter.Current().ID)
	}

	return nil, errors.Wrap(iter.Err(), "unable to get access application by domain")
}

func (a *API) FindFirstAccessApplicationOfType(ctx context.Context, app_type string) (*zero_trust.AccessApplicationGetResponse, error) {
	//
	iter := a.client.ZeroTrust.Access.Applications.ListAutoPaging(ctx, zero_trust.AccessApplicationListParams{
		AccountID: cloudflare.F(a.CFAccountID),
	})

	//
	for iter.Next() {
		current := iter.Current()
		if current.Type != app_type {
			// keep searching until we find correct type
			continue
		}
		return a.AccessApplication(ctx, current.ID)
	}

	return nil, errors.Wrap(iter.Err(), "unable to get access application of type")
}

func (a *API) AccessApplication(ctx context.Context, accessApplicationID string) (*zero_trust.AccessApplicationGetResponse, error) {
	//
	cfApp, err := a.client.ZeroTrust.Access.Applications.Get(ctx, accessApplicationID, zero_trust.AccessApplicationGetParams{
		AccountID: cloudflare.F(a.CFAccountID),
	})

	return cfApp, errors.Wrap(err, "unable to get access application")
}

//nolint:cyclop
func (a *API) CreateAccessApplication(
	ctx context.Context,
	app *v4alpha1.CloudflareAccessApplication, //nolint:varnamelen
) (*zero_trust.AccessApplicationGetResponse, error) {
	//
	var cfApp *zero_trust.AccessApplicationNewResponse
	var err error

	//
	switch app.Spec.Type {
	case string(zero_trust.ApplicationTypeSelfHosted):
		{
			body := zero_trust.AccessApplicationNewParamsBodySelfHostedApplication{
				Name:            cloudflare.F(app.Spec.Name),
				Domain:          cloudflare.F(app.Spec.Domain),
				AllowedIdPs:     cloudflare.F(app.Spec.AllowedIdps),
				Policies:        cloudflare.F(p_new_SH(app.Status.ReusablePolicyIDs)),
				SessionDuration: cloudflare.F(app.Spec.SessionDuration),
				LogoURL:         cloudflare.F(app.Spec.LogoURL),
			}
			if app.Spec.AppLauncherVisible != nil {
				body.AppLauncherVisible = cloudflare.Bool(*app.Spec.AppLauncherVisible)
			}
			if app.Spec.AutoRedirectToIdentity != nil {
				body.AutoRedirectToIdentity = cloudflare.Bool(*app.Spec.AutoRedirectToIdentity)
			}
			if app.Spec.EnableBindingCookie != nil {
				body.EnableBindingCookie = cloudflare.Bool(*app.Spec.EnableBindingCookie)
			}
			if app.Spec.HTTPOnlyCookieAttribute != nil {
				body.HTTPOnlyCookieAttribute = cloudflare.Bool(*app.Spec.HTTPOnlyCookieAttribute)
			}

			cfApp, err = a.client.ZeroTrust.Access.Applications.New(ctx, zero_trust.AccessApplicationNewParams{
				AccountID: cloudflare.F(a.CFAccountID),
				Body:      body,
			})
		}
	case string(zero_trust.ApplicationTypeWARP):
		{
			body := zero_trust.AccessApplicationNewParamsBodyDeviceEnrollmentPermissionsApplication{
				Type:               cloudflare.F(zero_trust.ApplicationType(app.Spec.Type)),
				AllowedIdPs:        cloudflare.F(app.Spec.AllowedIdps),
				Policies:           cloudflare.F(p_new_DEP(app.Status.ReusablePolicyIDs)),
				SessionDuration:    cloudflare.F(app.Spec.SessionDuration),
				AppLauncherLogoURL: cloudflare.F(app.Spec.LogoURL),
			}
			if app.Spec.AutoRedirectToIdentity != nil {
				body.AutoRedirectToIdentity = cloudflare.Bool(*app.Spec.AutoRedirectToIdentity)
			}

			cfApp, err = a.client.ZeroTrust.Access.Applications.New(ctx, zero_trust.AccessApplicationNewParams{
				AccountID: cloudflare.F(a.CFAccountID),
				Body:      body,
			})
		}
	case string(zero_trust.ApplicationTypeAppLauncher):
		{
			body := zero_trust.AccessApplicationNewParamsBodyAppLauncherApplication{
				Type:               cloudflare.F(zero_trust.ApplicationType(app.Spec.Type)),
				AllowedIdPs:        cloudflare.F(app.Spec.AllowedIdps),
				Policies:           cloudflare.F(p_new_AL(app.Status.ReusablePolicyIDs)),
				SessionDuration:    cloudflare.F(app.Spec.SessionDuration),
				AppLauncherLogoURL: cloudflare.F(app.Spec.LogoURL),
			}
			if app.Spec.AutoRedirectToIdentity != nil {
				body.AutoRedirectToIdentity = cloudflare.Bool(*app.Spec.AutoRedirectToIdentity)
			}

			cfApp, err = a.client.ZeroTrust.Access.Applications.New(ctx, zero_trust.AccessApplicationNewParams{
				AccountID: cloudflare.F(a.CFAccountID),
				Body:      body,
			})
		}
	default:
		{
			return nil, errors.Errorf("Unhandled application creation for '%s' app type. Contact the developers.", app.Spec.Type)
		}
	}

	if err != nil {
		return nil, errors.Wrap(err, "unable to create access application")
	}

	//
	if a.optionalTracer != nil {
		a.optionalTracer.ApplicationInserted(cfApp.ID)
	}

	return a.AccessApplication(ctx, cfApp.ID)
}

func (a *API) UpdateAccessApplication(
	ctx context.Context,
	app *v4alpha1.CloudflareAccessApplication, //nolint:varnamelen
) (*zero_trust.AccessApplicationGetResponse, error) {
	//
	var cfApp *zero_trust.AccessApplicationUpdateResponse
	var err error

	//
	switch app.Spec.Type {
	case string(zero_trust.ApplicationTypeSelfHosted):
		{
			body := zero_trust.AccessApplicationUpdateParamsBodySelfHostedApplication{
				Name:            cloudflare.F(app.Spec.Name),
				Domain:          cloudflare.F(app.Spec.Domain),
				AllowedIdPs:     cloudflare.F(app.Spec.AllowedIdps),
				Policies:        cloudflare.F(p_update_SH(app.Status.ReusablePolicyIDs)),
				SessionDuration: cloudflare.F(app.Spec.SessionDuration),
				LogoURL:         cloudflare.F(app.Spec.LogoURL),
			}
			if app.Spec.AppLauncherVisible != nil {
				body.AppLauncherVisible = cloudflare.Bool(*app.Spec.AppLauncherVisible)
			}
			if app.Spec.AutoRedirectToIdentity != nil {
				body.AutoRedirectToIdentity = cloudflare.Bool(*app.Spec.AutoRedirectToIdentity)
			}
			if app.Spec.EnableBindingCookie != nil {
				body.EnableBindingCookie = cloudflare.Bool(*app.Spec.EnableBindingCookie)
			}
			if app.Spec.HTTPOnlyCookieAttribute != nil {
				body.HTTPOnlyCookieAttribute = cloudflare.Bool(*app.Spec.HTTPOnlyCookieAttribute)
			}

			cfApp, err = a.client.ZeroTrust.Access.Applications.Update(ctx, app.Status.AccessApplicationID,
				zero_trust.AccessApplicationUpdateParams{
					AccountID: cloudflare.F(a.CFAccountID),
					Body:      body,
				},
			)
		}
	case string(zero_trust.ApplicationTypeWARP):
		{
			cfApp, err = a.client.ZeroTrust.Access.Applications.Update(ctx, app.Status.AccessApplicationID,
				zero_trust.AccessApplicationUpdateParams{
					AccountID: cloudflare.F(a.CFAccountID),
					Body: zero_trust.AccessApplicationUpdateParamsBodyDeviceEnrollmentPermissionsApplication{
						AllowedIdPs:            cloudflare.F(app.Spec.AllowedIdps),
						AutoRedirectToIdentity: cloudflare.F(*app.Spec.AutoRedirectToIdentity),
						Policies:               cloudflare.F(p_update_DEP(app.Status.ReusablePolicyIDs)),
						SessionDuration:        cloudflare.F(app.Spec.SessionDuration),
						AppLauncherLogoURL:     cloudflare.F(app.Spec.LogoURL),
					},
				},
			)
		}
	case string(zero_trust.ApplicationTypeAppLauncher):
		{
			cfApp, err = a.client.ZeroTrust.Access.Applications.Update(ctx, app.Status.AccessApplicationID,
				zero_trust.AccessApplicationUpdateParams{
					AccountID: cloudflare.F(a.CFAccountID),
					Body: zero_trust.AccessApplicationUpdateParamsBodyAppLauncherApplication{
						AllowedIdPs:            cloudflare.F(app.Spec.AllowedIdps),
						AutoRedirectToIdentity: cloudflare.F(*app.Spec.AutoRedirectToIdentity),
						Policies:               cloudflare.F(p_update_AL(app.Status.ReusablePolicyIDs)),
						SessionDuration:        cloudflare.F(app.Spec.SessionDuration),
						AppLauncherLogoURL:     cloudflare.F(app.Spec.LogoURL),
					},
				},
			)
		}
	default:
		{
			return nil, errors.Errorf("Unhandled application update for '%s' app type. Contact the developers.", app.Spec.Type)
		}
	}

	if err != nil {
		return nil, errors.Wrap(err, "unable to update access application")
	}

	return a.AccessApplication(ctx, cfApp.ID)
}

func (a *API) DeleteAccessApplication(ctx context.Context, appID string) error {
	//
	_, err := a.client.ZeroTrust.Access.Applications.Delete(ctx, appID, zero_trust.AccessApplicationDeleteParams{
		AccountID: cloudflare.F(a.CFAccountID),
	})

	//
	if a.optionalTracer != nil && err == nil {
		a.optionalTracer.ApplicationDeleted(appID)
	}

	return errors.Wrap(err, "unable to delete access application")
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
	rp, err := a.client.ZeroTrust.Access.Policies.New(ctx, zero_trust.AccessPolicyNewParams{ //nolint:varnamelen
		AccountID: cloudflare.F(a.CFAccountID),
		Decision:  cloudflare.F(zero_trust.Decision(arp.Spec.Decision)),
		Name:      cloudflare.F(arp.Spec.Name),
		Include:   cloudflare.F(arp.Spec.Include.ToAccessRuleParams(arp.Status.ResolvedIdpsFromRefs.Include)),
		Exclude:   cloudflare.F(arp.Spec.Exclude.ToAccessRuleParams(arp.Status.ResolvedIdpsFromRefs.Exclude)),
		Require:   cloudflare.F(arp.Spec.Require.ToAccessRuleParams(arp.Status.ResolvedIdpsFromRefs.Require)),
	})

	if err != nil {
		return nil, errors.Wrap(err, "unable to create access reusable policy")
	}

	//
	if a.optionalTracer != nil {
		a.optionalTracer.ReusablePolicyInserted(rp.ID)
	}

	return a.AccessReusablePolicy(ctx, rp.ID)
}

func (a *API) UpdateAccessReusablePolicy(ctx context.Context, arp *v4alpha1.CloudflareAccessReusablePolicy) error {
	_, err := a.client.ZeroTrust.Access.Policies.Update(ctx, arp.Status.AccessReusablePolicyID, zero_trust.AccessPolicyUpdateParams{
		AccountID: cloudflare.F(a.CFAccountID),
		Include:   cloudflare.F(arp.Spec.Include.ToAccessRuleParams(arp.Status.ResolvedIdpsFromRefs.Include)),
		Exclude:   cloudflare.F(arp.Spec.Exclude.ToAccessRuleParams(arp.Status.ResolvedIdpsFromRefs.Exclude)),
		Require:   cloudflare.F(arp.Spec.Require.ToAccessRuleParams(arp.Status.ResolvedIdpsFromRefs.Require)),
	})

	return errors.Wrap(err, "unable to update access reusable policy")
}

func (a *API) DeleteAccessReusablePolicy(ctx context.Context, policyID string) error {

	_, err := a.client.ZeroTrust.Access.Policies.Delete(ctx, policyID, zero_trust.AccessPolicyDeleteParams{
		AccountID: cloudflare.F(a.CFAccountID),
	})

	//
	if a.optionalTracer != nil && err == nil {
		a.optionalTracer.ReusablePolicyDeleted(policyID)
	}

	return errors.Wrap(err, "unable to delete access reusable policy")
}

//
//
//

func (a *API) AccessServiceToken(ctx context.Context, tokenId string) (*cftypes.ExtendedServiceToken, error) {

	token, err := a.client.ZeroTrust.Access.ServiceTokens.Get(ctx, tokenId, zero_trust.AccessServiceTokenGetParams{
		AccountID: cloudflare.F(a.CFAccountID),
	})

	if err != nil {
		return nil, errors.Wrap(err, "unable to get access service token")
	}

	return &cftypes.ExtendedServiceToken{ServiceToken: *token}, nil
}

func (a *API) AccessServiceTokens(ctx context.Context) (*[]cftypes.ExtendedServiceToken, error) {

	iter := a.client.ZeroTrust.Access.ServiceTokens.ListAutoPaging(ctx, zero_trust.AccessServiceTokenListParams{
		AccountID: cloudflare.F(a.CFAccountID),
	})

	extendedTokens := []cftypes.ExtendedServiceToken{}
	for iter.Next() {
		extendedTokens = append(extendedTokens, cftypes.ExtendedServiceToken{
			ServiceToken: iter.Current(),
		})
	}

	return &extendedTokens, errors.Wrap(iter.Err(), "unable to get access service tokens")
}

func (a *API) CreateAccessServiceToken(ctx context.Context, token cftypes.ExtendedServiceToken) (*cftypes.ExtendedServiceToken, error) {
	//
	res, err := a.client.ZeroTrust.Access.ServiceTokens.New(ctx, zero_trust.AccessServiceTokenNewParams{
		AccountID: cloudflare.F(a.CFAccountID),
		Name:      cloudflare.F(token.Name),
	})

	if err != nil {
		return nil, errors.Wrap(err, "unable to create access service token")
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

	//
	if a.optionalTracer != nil {
		a.optionalTracer.ServiceTokenInserted(sToken.ID)
	}

	return &extendedToken, errors.Wrap(err, "unable to create access service token")
}

func (a *API) DeleteAccessServiceToken(ctx context.Context, tokenID string) error {

	_, err := a.client.ZeroTrust.Access.ServiceTokens.Delete(ctx, tokenID, zero_trust.AccessServiceTokenDeleteParams{
		AccountID: cloudflare.F(a.CFAccountID),
	})

	//
	if a.optionalTracer != nil && err == nil {
		a.optionalTracer.ServiceTokenDeleted(tokenID)
	}

	return errors.Wrap(err, "unable to delete access service token")
}
