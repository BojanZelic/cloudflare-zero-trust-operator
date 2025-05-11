package cfapi

//
// These API call are most probably used only for tests
//

import (
	"context"

	cloudflare "github.com/cloudflare/cloudflare-go/v4"
	"github.com/cloudflare/cloudflare-go/v4/shared"
	"github.com/cloudflare/cloudflare-go/v4/zero_trust"
	"github.com/pkg/errors"
)

func (a *API) AccessGroups(ctx context.Context) (*[]zero_trust.AccessGroupListResponse, error) {
	//
	iter := a.client.ZeroTrust.Access.Groups.ListAutoPaging(ctx, zero_trust.AccessGroupListParams{
		AccountID: cloudflare.F(a.CFAccountID),
	})

	cfAccessGroups := []zero_trust.AccessGroupListResponse{}
	for iter.Next() {
		cfAccessGroups = append(cfAccessGroups, iter.Current())
	}

	//
	return &cfAccessGroups, errors.Wrap(iter.Err(), "unable to get access groups")
}

func (a *API) AccessApplications(ctx context.Context) (*[]zero_trust.AccessApplicationListResponse, error) {
	//
	iter := a.client.ZeroTrust.Access.Applications.ListAutoPaging(ctx, zero_trust.AccessApplicationListParams{
		AccountID: cloudflare.F(a.CFAccountID),
	})

	apps := []zero_trust.AccessApplicationListResponse{}
	for iter.Next() {
		apps = append(apps, iter.Current())
	}

	return &apps, errors.Wrap(iter.Err(), "unable to get access applications")
}

func (a *API) AccessReusablesPolicies(ctx context.Context) (*[]zero_trust.AccessPolicyListResponse, error) {
	//
	iter := a.client.ZeroTrust.Access.Policies.ListAutoPaging(ctx, zero_trust.AccessPolicyListParams{
		AccountID: cloudflare.F(a.CFAccountID),
	})

	cfAccessRP := []zero_trust.AccessPolicyListResponse{}
	for iter.Next() {
		if !iter.Current().Reusable {
			continue
		}
		cfAccessRP = append(cfAccessRP, iter.Current())
	}

	//
	return &cfAccessRP, errors.Wrap(iter.Err(), "unable to get access reusable policies")
}

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
	return &idProviders, errors.Wrap(iter.Err(), "unable to get identity providers")
}

//
// TODO(maintainer) Can those below be turned into a single generic func ?
//

func p_new_AL(ids []string) (out []zero_trust.AccessApplicationNewParamsBodyAppLauncherApplicationPolicyUnion) {
	for _, s := range ids {
		out = append(out, shared.UnionString(s))
	}
	return
}

func p_new_DEP(ids []string) (out []zero_trust.AccessApplicationNewParamsBodyDeviceEnrollmentPermissionsApplicationPolicyUnion) {
	for _, s := range ids {
		out = append(out, shared.UnionString(s))
	}
	return
}

func p_new_SH(ids []string) (out []zero_trust.AccessApplicationNewParamsBodySelfHostedApplicationPolicyUnion) {
	for _, s := range ids {
		out = append(out, shared.UnionString(s))
	}
	return
}

///

func p_update_AL(ids []string) (out []zero_trust.AccessApplicationUpdateParamsBodyAppLauncherApplicationPolicyUnion) {
	for _, s := range ids {
		out = append(out, shared.UnionString(s))
	}
	return
}

func p_update_DEP(ids []string) (out []zero_trust.AccessApplicationUpdateParamsBodyDeviceEnrollmentPermissionsApplicationPolicyUnion) {
	for _, s := range ids {
		out = append(out, shared.UnionString(s))
	}
	return
}

func p_update_SH(ids []string) (out []zero_trust.AccessApplicationUpdateParamsBodySelfHostedApplicationPolicyUnion) {
	for _, s := range ids {
		out = append(out, shared.UnionString(s))
	}
	return
}
