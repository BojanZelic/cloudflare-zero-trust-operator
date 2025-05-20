package cfapi

//
// These API call are most probably used only for tests
//

import (
	"context"

	"github.com/Southclaws/fault"
	"github.com/Southclaws/fault/fctx"
	"github.com/Southclaws/fault/fmsg"
	"github.com/bojanzelic/cloudflare-zero-trust-operator/api/v4alpha1"
	cloudflare "github.com/cloudflare/cloudflare-go/v4"
	"github.com/cloudflare/cloudflare-go/v4/zero_trust"
	"github.com/cloudflare/cloudflare-go/v4/zones"
)

// To print available identity providers at boot
func (a *API) IdentityProviders(ctx context.Context) (*[]zero_trust.IdentityProviderListResponse, error) {
	//
	iter := a.client.ZeroTrust.IdentityProviders.ListAutoPaging(ctx, zero_trust.IdentityProviderListParams{
		AccountID: cloudflare.F(a.CFAccountID),
	})

	idProviders := []zero_trust.IdentityProviderListResponse{}
	for iter.Next() {
		idProviders = append(idProviders, iter.Current())
	}

	//
	return &idProviders, fault.Wrap(iter.Err(), fmsg.With("unable to get identity providers"))
}

// Testing purposes only
func (a *API) IsDomainOwned(ctx context.Context, domainName string) (bool, error) {
	//
	iter := a.client.Zones.ListAutoPaging(ctx, zones.ZoneListParams{})

	for iter.Next() {
		zone := iter.Current()
		if zone.Name == domainName {
			return true, nil
		}
	}

	return false, fault.Wrap(iter.Err(),
		fmsg.With("unable to determine domain ownership"),
		fctx.With(ctx,
			"searchedDomain", domainName,
		),
	)
}

// DEV ONLY !
// @dev enforce deletion on appID; outside of tests, should prefer [DeleteOrResetAccessApplication]
func (a *API) DeleteGenericAccessApplication(ctx context.Context, appID string) error {
	return a.deleteAccessApplication(ctx, appID)
}

//
//
//

// will remove identified policy UUIDs bound to a CRD (and only thoses) of the CloudFlare remote counterpart
// Functionnally similar to a deletion, but for account-wide application - which cannot be deleted
// @dev helper to be used in [cfapi] internally
func (a *API) resetAccessApplication(ctx context.Context,
	appToReset *v4alpha1.CloudflareAccessApplication,
	appToPreserveFrom *zero_trust.AccessApplicationGetResponse,
) error {
	//
	orderedUUIDs, err := GetOrderedPolicyUUIDs(appToPreserveFrom)
	if err != nil {
		return fault.Wrap(err,
			fmsg.WithDesc("Issue while preparing application reset", "Unable to produce list of policies from remote resource"),
		)
	}

	//
	policyIDsToResetTo := difference(orderedUUIDs, appToReset.Status.ReusablePolicyIDs)

	return a._resetAccessApplication(ctx,
		appToReset.GetCloudflareUUID(),
		appToReset.Spec.Type,
		policyIDsToResetTo,
	)
}

// @dev test-only
func (a *API) RestoreAccessApplicationTo(ctx context.Context, appToResetTo *zero_trust.AccessApplicationGetResponse) error {
	//
	policyIDsToResetTo, err := GetOrderedPolicyUUIDs(appToResetTo)
	if err != nil {
		return fault.Wrap(err,
			fmsg.WithDesc("Issue while preparing application reset", "Unable to produce list of policies from remote resource"),
		)
	}

	//
	return a._resetAccessApplication(ctx,
		appToResetTo.ID,
		appToResetTo.Type,
		policyIDsToResetTo,
	)
}

// Update access app by trimming resolved policy UUIDs from remote resource.
func (a *API) _resetAccessApplication(ctx context.Context,
	appIDToReset string,
	appTypeToReset string,
	policyIDsToResetTo []string,
) error {

	//
	var body zero_trust.AccessApplicationUpdateParamsBodyUnion
	switch appTypeToReset {
	case string(zero_trust.ApplicationTypeAppLauncher):
		{
			body = zero_trust.AccessApplicationUpdateParamsBodyAppLauncherApplication{
				Type:     cloudflare.F(zero_trust.ApplicationTypeAppLauncher), // always required
				Policies: cloudflare.F(p_update_AL(policyIDsToResetTo)),
			}
		}
	case string(zero_trust.ApplicationTypeWARP):
		{
			body = zero_trust.AccessApplicationUpdateParamsBodyDeviceEnrollmentPermissionsApplication{
				Type:     cloudflare.F(zero_trust.ApplicationTypeWARP), // always required
				Policies: cloudflare.F(p_update_DEP(policyIDsToResetTo)),
			}
		}
	default:
		{
			return fault.Newf("Unhandled application reset for '%s' app type. Contact the developers.", appTypeToReset)
		}
	}

	//
	_, err := a.client.ZeroTrust.Access.Applications.Update(ctx,
		appIDToReset,
		zero_trust.AccessApplicationUpdateParams{
			AccountID: cloudflare.F(a.CFAccountID),
			Body:      body,
		},
	)

	return a.wrapPrettyForAPI(err)
}

//
//
//

// basic delete of application
func (a *API) deleteAccessApplication(ctx context.Context, appID string) error {
	//
	_, err := a.client.ZeroTrust.Access.Applications.Delete(ctx, appID, zero_trust.AccessApplicationDeleteParams{
		AccountID: cloudflare.F(a.CFAccountID),
	})

	//
	if a.optionalTracer != nil && err == nil {
		a.optionalTracer.ApplicationDeleted(appID)
	}

	return a.wrapPrettyForAPI(err)
}
