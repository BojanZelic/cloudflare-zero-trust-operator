package cfapi

import (
	"context"

	"github.com/bojanzelic/cloudflare-zero-trust-operator/api/v1alpha1"
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
		empty := zero_trust.AccessApplicationGetResponse{}
		return &empty, errors.Wrap(iter.Err(), "unable to get access applications")
	}

	return a.AccessApplication(ctx, iter.Current().ID)
}

func (a *API) AccessApplication(ctx context.Context, accessApplicationID string) (*zero_trust.AccessApplicationGetResponse, error) {
	//
	cfAG, err := a.client.ZeroTrust.Access.Applications.Get(ctx, accessApplicationID, zero_trust.AccessApplicationGetParams{
		AccountID: cloudflare.F(a.CFAccountID),
	})

	return cfAG, errors.Wrap(err, "unable to get access application")
}

func (a *API) CreateAccessApplication(
	ctx context.Context,
	ag *v1alpha1.CloudflareAccessApplication, //nolint:varnamelen
) (*zero_trust.AccessApplicationGetResponse, error) {
	//
	cfAG, err := a.client.ZeroTrust.Access.Applications.New(ctx, zero_trust.AccessApplicationNewParams{
		AccountID: cloudflare.F(a.CFAccountID),
		Body: zero_trust.AccessApplicationNewParamsBody{
			Name:                   cloudflare.F(ag.Spec.Name),
			Domain:                 cloudflare.F(ag.Spec.Domain),
			Type:                   cloudflare.F(ag.Spec.Type),
			AppLauncherVisible:     cloudflare.F(*ag.Spec.AppLauncherVisible),
			AllowedIdPs:            cloudflare.F(interface{}(ag.Spec.AllowedIdps)),
			AutoRedirectToIdentity: cloudflare.F(*ag.Spec.AutoRedirectToIdentity),
			// Policies: , // TODO: maybe handle old legacy policies
			SessionDuration:         cloudflare.F(ag.Spec.SessionDuration),
			EnableBindingCookie:     cloudflare.F(*ag.Spec.EnableBindingCookie),
			HTTPOnlyCookieAttribute: cloudflare.F(*ag.Spec.HTTPOnlyCookieAttribute),
			LogoURL:                 cloudflare.F(ag.Spec.LogoURL),
		},
	})

	if err != nil {
		dummy := zero_trust.AccessApplicationGetResponse{}
		return &dummy, errors.Wrap(err, "unable to create access applications")
	}

	return a.AccessApplication(ctx, cfAG.ID)
}

func (a *API) UpdateAccessApplication(
	ctx context.Context,
	ag zero_trust.AccessApplicationGetResponse, //nolint:varnamelen
) (*zero_trust.AccessApplicationGetResponse, error) {
	//
	cfAG, err := a.client.ZeroTrust.Access.Applications.Update(ctx, ag.ID, zero_trust.AccessApplicationUpdateParams{
		AccountID: cloudflare.F(a.CFAccountID),
		Body: zero_trust.AccessApplicationUpdateParamsBody{
			Name:                   cloudflare.F(ag.Name),
			Domain:                 cloudflare.F(ag.Domain),
			Type:                   cloudflare.F(ag.Type),
			AppLauncherVisible:     cloudflare.F(ag.AppLauncherVisible),
			AllowedIdPs:            cloudflare.F(ag.AllowedIdPs),
			AutoRedirectToIdentity: cloudflare.F(ag.AutoRedirectToIdentity),
			// Policies: , // TODO: maybe handle old legacy policies
			SessionDuration:         cloudflare.F(ag.SessionDuration),
			EnableBindingCookie:     cloudflare.F(ag.EnableBindingCookie),
			HTTPOnlyCookieAttribute: cloudflare.F(ag.HTTPOnlyCookieAttribute),
			LogoURL:                 cloudflare.F(ag.LogoURL),
		},
	})

	if err != nil {
		dummy := zero_trust.AccessApplicationGetResponse{}
		return &dummy, errors.Wrap(err, "unable to update access applications")
	}

	return a.AccessApplication(ctx, cfAG.ID)
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

func (a *API) AccessApplicationPolicies(ctx context.Context, appID string) (cfcollections.AccessApplicationPolicyCollection, error) {
	//
	iter := a.client.ZeroTrust.Access.Applications.Policies.ListAutoPaging(ctx, appID, zero_trust.AccessApplicationPolicyListParams{
		AccountID: cloudflare.F(a.CFAccountID),
	})

	policiesCollection := cfcollections.AccessApplicationPolicyCollection{}
	for iter.Next() {
		policiesCollection = append(policiesCollection, iter.Current())
	}

	return policiesCollection, errors.Wrap(iter.Err(), "unable to get access Policies")
}

func (a *API) CreateAccessApplicationPolicies(ctx context.Context, appID string, lap zero_trust.AccessApplicationPolicyListResponse) error {
	//
	approvalGroups := []zero_trust.ApprovalGroupParam{}
	for _, r := range lap.ApprovalGroups {
		approvalGroups = append(approvalGroups, zero_trust.ApprovalGroupParam{
			ApprovalsNeeded: cloudflare.F(r.ApprovalsNeeded),
			EmailAddresses:  cloudflare.F(r.EmailAddresses),
			EmailListUUID:   cloudflare.F(r.EmailListUUID),
		})
	}

	_, err := a.client.ZeroTrust.Access.Applications.Policies.New(ctx, appID, zero_trust.AccessApplicationPolicyNewParams{
		AccountID:                    cloudflare.F(a.CFAccountID),
		Precedence:                   cloudflare.F(lap.Precedence),
		IsolationRequired:            cloudflare.F(lap.IsolationRequired),
		SessionDuration:              cloudflare.F(lap.SessionDuration),
		PurposeJustificationRequired: cloudflare.F(lap.PurposeJustificationRequired),
		PurposeJustificationPrompt:   cloudflare.F(lap.PurposeJustificationPrompt),
		ApprovalRequired:             cloudflare.F(lap.ApprovalRequired),
		ApprovalGroups:               cloudflare.F(approvalGroups),
	})

	// TODO create policy

	return errors.Wrap(err, "unable to create access Policy")
}

func (a *API) UpdateAccessApplicationPolicy(ctx context.Context, appID string, lap zero_trust.AccessApplicationPolicyListResponse) error {
	//
	approvalGroups := []zero_trust.ApprovalGroupParam{}
	for _, r := range lap.ApprovalGroups {
		approvalGroups = append(approvalGroups, zero_trust.ApprovalGroupParam{
			ApprovalsNeeded: cloudflare.F(r.ApprovalsNeeded),
			EmailAddresses:  cloudflare.F(r.EmailAddresses),
			EmailListUUID:   cloudflare.F(r.EmailListUUID),
		})
	}

	_, err := a.client.ZeroTrust.Access.Applications.Policies.Update(ctx, appID, lap.ID, zero_trust.AccessApplicationPolicyUpdateParams{
		AccountID:                    cloudflare.F(a.CFAccountID),
		Precedence:                   cloudflare.F(lap.Precedence),
		IsolationRequired:            cloudflare.F(lap.IsolationRequired),
		SessionDuration:              cloudflare.F(lap.SessionDuration),
		PurposeJustificationRequired: cloudflare.F(lap.PurposeJustificationRequired),
		PurposeJustificationPrompt:   cloudflare.F(lap.PurposeJustificationPrompt),
		ApprovalRequired:             cloudflare.F(lap.ApprovalRequired),
		ApprovalGroups:               cloudflare.F(approvalGroups),
	})

	// TODO create policy

	return errors.Wrap(err, "unable to update access Policy")
}

func (a *API) DeleteAccessApplicationPolicy(ctx context.Context, appID string, policyID string) error {

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
