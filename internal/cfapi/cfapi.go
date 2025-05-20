package cfapi

import (
	"context"

	"github.com/Southclaws/fault"
	"github.com/Southclaws/fault/fctx"
	"github.com/Southclaws/fault/fmsg"
	"github.com/bojanzelic/cloudflare-zero-trust-operator/api/v4alpha1"
	"github.com/bojanzelic/cloudflare-zero-trust-operator/internal/cftypes"
	cloudflare "github.com/cloudflare/cloudflare-go/v4"
	"github.com/cloudflare/cloudflare-go/v4/zero_trust"
)

//
// Access Group
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
	return nil, a.wrapPrettyForAPI(iter.Err())
}

func (a *API) AccessGroup(ctx context.Context, accessGroupID string) (*zero_trust.AccessGroupGetResponse, error) {
	//
	cfAG, err := a.client.ZeroTrust.Access.Groups.Get(ctx, accessGroupID, zero_trust.AccessGroupGetParams{
		AccountID: cloudflare.F(a.CFAccountID),
	})

	return cfAG, a.wrapPrettyForAPI(err)
}

func (a *API) CreateAccessGroup(ctx context.Context, group *v4alpha1.CloudflareAccessGroup) (*zero_trust.AccessGroupGetResponse, error) {
	//
	params := zero_trust.AccessGroupNewParams{
		AccountID: cloudflare.F(a.CFAccountID),
		Name:      cloudflare.F(group.Spec.Name),
		Include:   cloudflare.F(group.Spec.Include.ToAccessRuleParams(group.Status.ResolvedIdpsFromRefs.Include)),
		Exclude:   cloudflare.F(group.Spec.Exclude.ToAccessRuleParams(group.Status.ResolvedIdpsFromRefs.Exclude)),
		Require:   cloudflare.F(group.Spec.Require.ToAccessRuleParams(group.Status.ResolvedIdpsFromRefs.Require)),
	}

	//
	insert, err := a.client.ZeroTrust.Access.Groups.New(ctx, params)
	if err != nil {
		return nil, a.wrapPrettyForAPI(err)
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
	params := zero_trust.AccessGroupUpdateParams{
		AccountID: cloudflare.F(a.CFAccountID),
		Name:      cloudflare.F(group.Spec.Name),
		Include:   cloudflare.F(group.Spec.Include.ToAccessRuleParams(group.Status.ResolvedIdpsFromRefs.Include)),
		Exclude:   cloudflare.F(group.Spec.Exclude.ToAccessRuleParams(group.Status.ResolvedIdpsFromRefs.Exclude)),
		Require:   cloudflare.F(group.Spec.Require.ToAccessRuleParams(group.Status.ResolvedIdpsFromRefs.Require)),
	}

	//
	_, err := a.client.ZeroTrust.Access.Groups.Update(ctx, group.GetCloudflareUUID(), params)
	return a.wrapPrettyForAPI(err)
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

	return a.wrapPrettyForAPI(err)
}

//
// Access Application
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

	return nil, a.wrapPrettyForAPI(iter.Err())
}

// not finding app type would probably not produce an error
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

	//
	return nil, a.wrapPrettyForAPI(iter.Err())
}

func (a *API) AccessApplication(ctx context.Context, accessApplicationID string) (*zero_trust.AccessApplicationGetResponse, error) {
	//
	cfApp, err := a.client.ZeroTrust.Access.Applications.Get(ctx, accessApplicationID, zero_trust.AccessApplicationGetParams{
		AccountID: cloudflare.F(a.CFAccountID),
	})

	return cfApp, a.wrapPrettyForAPI(err)
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
				Type:            cloudflare.F(app.Spec.Type),
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
			return nil, fault.Newf("Unhandled application creation for '%s' app type. Contact the developers.", app.Spec.Type)
		}
	}

	if err != nil {
		return nil, a.wrapPrettyForAPI(err)
	}

	//
	if a.optionalTracer != nil {
		a.optionalTracer.ApplicationInserted(cfApp.ID)
	}

	return a.AccessApplication(ctx, cfApp.ID)
}

//nolint:cyclop
func (a *API) UpdateAccessApplication(
	ctx context.Context,
	app *v4alpha1.CloudflareAccessApplication, //nolint:varnamelen
) (*zero_trust.AccessApplicationGetResponse, error) {
	//
	var cfAppResp *zero_trust.AccessApplicationUpdateResponse
	var err error
	appIDToUpdate := app.GetCloudflareUUID()

	//
	switch app.Spec.Type {
	case string(zero_trust.ApplicationTypeSelfHosted):
		{
			body := zero_trust.AccessApplicationUpdateParamsBodySelfHostedApplication{
				Type:            cloudflare.F(app.Spec.Type),
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

			cfAppResp, err = a.client.ZeroTrust.Access.Applications.Update(ctx, appIDToUpdate,
				zero_trust.AccessApplicationUpdateParams{
					AccountID: cloudflare.F(a.CFAccountID),
					Body:      body,
				},
			)
		}
	case string(zero_trust.ApplicationTypeWARP):
		{
			body := zero_trust.AccessApplicationUpdateParamsBodyDeviceEnrollmentPermissionsApplication{
				Type:               cloudflare.F(zero_trust.ApplicationTypeWARP), // always required
				AllowedIdPs:        cloudflare.F(app.Spec.AllowedIdps),
				Policies:           cloudflare.F(p_update_DEP(app.Status.ReusablePolicyIDs)),
				SessionDuration:    cloudflare.F(app.Spec.SessionDuration),
				AppLauncherLogoURL: cloudflare.F(app.Spec.LogoURL),
			}

			if app.Spec.AutoRedirectToIdentity != nil {
				body.AutoRedirectToIdentity = cloudflare.Bool(*app.Spec.AutoRedirectToIdentity)
			}

			cfAppResp, err = a.client.ZeroTrust.Access.Applications.Update(ctx, appIDToUpdate,
				zero_trust.AccessApplicationUpdateParams{
					AccountID: cloudflare.F(a.CFAccountID),
					Body:      body,
				},
			)
		}
	case string(zero_trust.ApplicationTypeAppLauncher):
		{
			body := zero_trust.AccessApplicationUpdateParamsBodyAppLauncherApplication{
				Type:               cloudflare.F(zero_trust.ApplicationTypeAppLauncher), // always required
				AllowedIdPs:        cloudflare.F(app.Spec.AllowedIdps),
				Policies:           cloudflare.F(p_update_AL(app.Status.ReusablePolicyIDs)),
				SessionDuration:    cloudflare.F(app.Spec.SessionDuration),
				AppLauncherLogoURL: cloudflare.F(app.Spec.LogoURL),
			}

			if app.Spec.AutoRedirectToIdentity != nil {
				body.AutoRedirectToIdentity = cloudflare.Bool(*app.Spec.AutoRedirectToIdentity)
			}

			cfAppResp, err = a.client.ZeroTrust.Access.Applications.Update(ctx, appIDToUpdate,
				zero_trust.AccessApplicationUpdateParams{
					AccountID: cloudflare.F(a.CFAccountID),
					Body:      body,
				},
			)
		}
	default:
		{
			return nil, fault.Newf("Unhandled application update for '%s' app type. Contact the developers.", app.Spec.Type)
		}
	}

	if err != nil {
		return nil, a.wrapPrettyForAPI(err)
	}

	return a.AccessApplication(ctx, cfAppResp.ID)
}

// @dev we cannot remove one-time special apps like "warp" or "app_launcher type", we need to reset those instead of deleting
func (a *API) DeleteOrResetAccessApplication(ctx context.Context, targetedApp *v4alpha1.CloudflareAccessApplication) error {
	switch targetedApp.Spec.Type {
	case string(zero_trust.ApplicationTypeSelfHosted):
		{
			return a.deleteAccessApplication(ctx, targetedApp.GetCloudflareUUID())
		}

	case string(zero_trust.ApplicationTypeAppLauncher),
		string(zero_trust.ApplicationTypeWARP):
		{
			//
			appToPreserveFrom, err := a.FindFirstAccessApplicationOfType(ctx, targetedApp.Spec.Type)
			if err != nil {
				return fault.Wrap(
					a.wrapPrettyForAPI(err),
					fmsg.WithDesc("Issue while preparing application reset", "Cannot find existing application of type"),
					fctx.With(ctx,
						"appType", targetedApp.Spec.Type,
					),
				)
			}

			//
			return a.resetAccessApplication(ctx, targetedApp, appToPreserveFrom)
		}
	}

	//
	return fault.Newf("Unhandled application deletion/reset for '%s' app type. Contact the developers.", targetedApp.Spec.Type)
}

//
// Access Reusable Policy
//

func (a *API) AccessReusablePolicy(ctx context.Context, policyID string) (*zero_trust.AccessPolicyGetResponse, error) {
	//
	cfApp, err := a.client.ZeroTrust.Access.Policies.Get(ctx, policyID, zero_trust.AccessPolicyGetParams{
		AccountID: cloudflare.F(a.CFAccountID),
	})

	return cfApp, a.wrapPrettyForAPI(err)
}

func (a *API) CreateAccessReusablePolicy(ctx context.Context, from *v4alpha1.CloudflareAccessReusablePolicy) (*zero_trust.AccessPolicyGetResponse, error) {
	//
	params := zero_trust.AccessPolicyNewParams{
		AccountID: cloudflare.F(a.CFAccountID),
		Decision:  cloudflare.F(zero_trust.Decision(from.Spec.Decision)),
		Name:      cloudflare.F(from.Spec.Name),
		Include:   cloudflare.F(from.Spec.Include.ToAccessRuleParams(from.Status.ResolvedIdpsFromRefs.Include)),
		Exclude:   cloudflare.F(from.Spec.Exclude.ToAccessRuleParams(from.Status.ResolvedIdpsFromRefs.Exclude)),
		Require:   cloudflare.F(from.Spec.Require.ToAccessRuleParams(from.Status.ResolvedIdpsFromRefs.Require)),
	}

	//
	arp, err := a.client.ZeroTrust.Access.Policies.New(ctx, params) //nolint:varnamelen
	if err != nil {
		return nil, a.wrapPrettyForAPI(err)
	}

	//
	if a.optionalTracer != nil {
		a.optionalTracer.ReusablePolicyInserted(arp.ID)
	}

	return a.AccessReusablePolicy(ctx, arp.ID)
}

func (a *API) UpdateAccessReusablePolicy(ctx context.Context, arp *v4alpha1.CloudflareAccessReusablePolicy) error {
	//
	params := zero_trust.AccessPolicyUpdateParams{
		AccountID: cloudflare.F(a.CFAccountID),
		Include:   cloudflare.F(arp.Spec.Include.ToAccessRuleParams(arp.Status.ResolvedIdpsFromRefs.Include)),
		Exclude:   cloudflare.F(arp.Spec.Exclude.ToAccessRuleParams(arp.Status.ResolvedIdpsFromRefs.Exclude)),
		Require:   cloudflare.F(arp.Spec.Require.ToAccessRuleParams(arp.Status.ResolvedIdpsFromRefs.Require)),
	}

	//
	_, err := a.client.ZeroTrust.Access.Policies.Update(ctx, arp.GetCloudflareUUID(), params)
	return a.wrapPrettyForAPI(err)
}

func (a *API) DeleteAccessReusablePolicy(ctx context.Context, policyID string) error {

	_, err := a.client.ZeroTrust.Access.Policies.Delete(ctx, policyID, zero_trust.AccessPolicyDeleteParams{
		AccountID: cloudflare.F(a.CFAccountID),
	})

	//
	if a.optionalTracer != nil && err == nil {
		a.optionalTracer.ReusablePolicyDeleted(policyID)
	}

	return a.wrapPrettyForAPI(err)
}

//
// Access Service Token
//

func (a *API) AccessServiceToken(ctx context.Context, tokenId string) (*cftypes.ExtendedServiceToken, error) {

	token, err := a.client.ZeroTrust.Access.ServiceTokens.Get(ctx, tokenId, zero_trust.AccessServiceTokenGetParams{
		AccountID: cloudflare.F(a.CFAccountID),
	})

	if err != nil {
		return nil, a.wrapPrettyForAPI(err)
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

	return &extendedTokens, a.wrapPrettyForAPI(iter.Err())
}

func (a *API) CreateAccessServiceToken(ctx context.Context, token cftypes.ExtendedServiceToken) (*cftypes.ExtendedServiceToken, error) {
	//
	res, err := a.client.ZeroTrust.Access.ServiceTokens.New(ctx, zero_trust.AccessServiceTokenNewParams{
		AccountID: cloudflare.F(a.CFAccountID),
		Name:      cloudflare.F(token.Name),
	})

	if err != nil {
		return nil, a.wrapPrettyForAPI(err)
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

	return &extendedToken, a.wrapPrettyForAPI(err)
}

func (a *API) DeleteAccessServiceToken(ctx context.Context, tokenID string) error {

	_, err := a.client.ZeroTrust.Access.ServiceTokens.Delete(ctx, tokenID, zero_trust.AccessServiceTokenDeleteParams{
		AccountID: cloudflare.F(a.CFAccountID),
	})

	//
	if a.optionalTracer != nil && err == nil {
		a.optionalTracer.ServiceTokenDeleted(tokenID)
	}

	return a.wrapPrettyForAPI(err)
}
