package cfapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/bojanzelic/cloudflare-zero-trust-operator/internal/cfcollections"
	"github.com/bojanzelic/cloudflare-zero-trust-operator/internal/cftypes"
	cloudflare "github.com/cloudflare/cloudflare-go/v6"
	"github.com/cloudflare/cloudflare-go/v6/option"
	"github.com/cloudflare/cloudflare-go/v6/zero_trust"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
)

const (
	cfErrAccessNotFound     = 11021
	cfErrAccessAppNotFound  = 12130
	errEmptyCloudflareCreds = "either CLOUDFLARE_API_TOKEN or (CLOUDFLARE_API_KEY and CLOUDFLARE_API_EMAIL) must be set"
)

type API struct {
	CFAccountID string
	client      *cloudflare.Client
}

func New(cfAPIToken string, cfAPIKey string, cfAPIEmail string, cfAccountID string) (*API, error) {
	if cfAPIToken == "" && (cfAPIKey == "" || cfAPIEmail == "") {
		return nil, errors.New(errEmptyCloudflareCreds)
	}

	var opts []option.RequestOption
	if cfAPIToken != "" {
		opts = append(opts,
			option.WithAPIToken(cfAPIToken),
			option.WithHeaderDel("X-Auth-Key"),
			option.WithHeaderDel("X-Auth-Email"),
			option.WithHeaderDel("X-Auth-User-Service-Key"),
		)
	} else {
		opts = append(opts,
			option.WithAPIKey(cfAPIKey),
			option.WithAPIEmail(cfAPIEmail),
			option.WithHeaderDel("authorization"),
			option.WithHeaderDel("X-Auth-User-Service-Key"),
		)
	}

	return &API{
		CFAccountID: cfAccountID,
		client:      cloudflare.NewClient(opts...),
	}, nil
}

// IsNotFoundError returns true when err comes from a Cloudflare 404 / resource-not-found API error.
func IsNotFoundError(err error) bool {
	if err == nil {
		return false
	}

	var apiErr *cloudflare.Error
	if errors.As(err, &apiErr) {
		if apiErr.StatusCode == http.StatusNotFound {
			return true
		}
		for _, e := range apiErr.Errors {
			if e.Code == cfErrAccessNotFound || e.Code == cfErrAccessAppNotFound {
				return true
			}
		}
	}

	return false
}

func parseTimeField(rawJSON string, field string) *time.Time {
	if rawJSON == "" {
		return nil
	}
	res := gjson.Get(rawJSON, field)
	if !res.Exists() || res.String() == "" {
		return nil
	}
	for _, layout := range []string{time.RFC3339Nano, time.RFC3339} {
		if parsed, err := time.Parse(layout, res.String()); err == nil {
			return &parsed
		}
	}

	return nil
}

func parseCreatedAndUpdated(rawJSON string) (*time.Time, *time.Time) {
	createdAt := parseTimeField(rawJSON, "created_at")
	updatedAt := parseTimeField(rawJSON, "updated_at")
	if createdAt == nil && updatedAt != nil {
		createdAt = updatedAt
	}
	if updatedAt == nil && createdAt != nil {
		updatedAt = createdAt
	}
	if createdAt == nil && updatedAt == nil {
		now := time.Now().UTC().Truncate(time.Second)

		return &now, &now
	}

	return createdAt, updatedAt
}

func parseRulesFromCloudflare(rules []zero_trust.AccessRule) []interface{} {
	out := make([]interface{}, 0, len(rules))
	for i := range rules {
		raw := rules[i].JSON.RawJSON()
		if raw == "" {
			continue
		}
		var item map[string]interface{}
		if err := json.Unmarshal([]byte(raw), &item); err == nil {
			out = append(out, item)
		}
	}

	return out
}

func toAccessRuleUnionParams(rules []interface{}) ([]zero_trust.AccessRuleUnionParam, bool) {
	typed := make([]zero_trust.AccessRuleUnionParam, 0, len(rules))
	for _, rule := range rules {
		u, ok := rule.(zero_trust.AccessRuleUnionParam)
		if !ok {
			return nil, false
		}
		typed = append(typed, u)
	}

	return typed, true
}

func parseAccessGroupFromRaw(id, name string, include, exclude, require []zero_trust.AccessRule, rawJSON string) cftypes.AccessGroup {
	createdAt, updatedAt := parseCreatedAndUpdated(rawJSON)

	return cftypes.AccessGroup{
		ID:        id,
		Name:      name,
		Include:   parseRulesFromCloudflare(include),
		Exclude:   parseRulesFromCloudflare(exclude),
		Require:   parseRulesFromCloudflare(require),
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}
}

func (a *API) AccessGroups(ctx context.Context) (cfcollections.AccessGroupCollection, error) {
	pager := a.client.ZeroTrust.Access.Groups.ListAutoPaging(ctx, zero_trust.AccessGroupListParams{
		AccountID: cloudflare.F(a.CFAccountID),
	})

	groups := cfcollections.AccessGroupCollection{}
	for pager.Next() {
		item := pager.Current()
		groups = append(groups, parseAccessGroupFromRaw(
			item.ID,
			item.Name,
			item.Include,
			item.Exclude,
			item.Require,
			item.JSON.RawJSON(),
		))
	}

	if err := pager.Err(); err != nil {
		return groups, errors.Wrap(err, "unable to get access groups")
	}

	return groups, nil
}

func (a *API) AccessGroup(ctx context.Context, accessGroupID string) (cftypes.AccessGroup, error) {
	res, err := a.client.ZeroTrust.Access.Groups.Get(ctx, accessGroupID, zero_trust.AccessGroupGetParams{
		AccountID: cloudflare.F(a.CFAccountID),
	})
	if err != nil {
		return cftypes.AccessGroup{}, errors.Wrap(err, "unable to get access group")
	}

	return parseAccessGroupFromRaw(
		res.ID,
		res.Name,
		res.Include,
		res.Exclude,
		res.Require,
		res.JSON.RawJSON(),
	), nil
}

func (a *API) CreateAccessGroup(ctx context.Context, ag cftypes.AccessGroup) (cftypes.AccessGroup, error) {
	params := zero_trust.AccessGroupNewParams{
		AccountID: cloudflare.F(a.CFAccountID),
		Name:      cloudflare.F(ag.Name),
	}
	if typed, ok := toAccessRuleUnionParams(ag.Include); ok {
		params.Include = cloudflare.F(typed)
	} else {
		params.Include = cloudflare.Raw[[]zero_trust.AccessRuleUnionParam](ag.Include)
	}
	if len(ag.Exclude) > 0 {
		if typed, ok := toAccessRuleUnionParams(ag.Exclude); ok {
			params.Exclude = cloudflare.F(typed)
		} else {
			params.Exclude = cloudflare.Raw[[]zero_trust.AccessRuleUnionParam](ag.Exclude)
		}
	}
	if len(ag.Require) > 0 {
		if typed, ok := toAccessRuleUnionParams(ag.Require); ok {
			params.Require = cloudflare.F(typed)
		} else {
			params.Require = cloudflare.Raw[[]zero_trust.AccessRuleUnionParam](ag.Require)
		}
	}

	res, err := a.client.ZeroTrust.Access.Groups.New(ctx, params)
	if err != nil {
		return cftypes.AccessGroup{}, errors.Wrap(err, "unable to create access groups")
	}

	return parseAccessGroupFromRaw(
		res.ID,
		res.Name,
		res.Include,
		res.Exclude,
		res.Require,
		res.JSON.RawJSON(),
	), nil
}

func (a *API) UpdateAccessGroup(ctx context.Context, ag cftypes.AccessGroup) (cftypes.AccessGroup, error) {
	params := zero_trust.AccessGroupUpdateParams{
		AccountID: cloudflare.F(a.CFAccountID),
		Name:      cloudflare.F(ag.Name),
	}
	if typed, ok := toAccessRuleUnionParams(ag.Include); ok {
		params.Include = cloudflare.F(typed)
	} else {
		params.Include = cloudflare.Raw[[]zero_trust.AccessRuleUnionParam](ag.Include)
	}
	if typed, ok := toAccessRuleUnionParams(ag.Exclude); ok {
		params.Exclude = cloudflare.F(typed)
	} else {
		params.Exclude = cloudflare.Raw[[]zero_trust.AccessRuleUnionParam](ag.Exclude)
	}
	if typed, ok := toAccessRuleUnionParams(ag.Require); ok {
		params.Require = cloudflare.F(typed)
	} else {
		params.Require = cloudflare.Raw[[]zero_trust.AccessRuleUnionParam](ag.Require)
	}

	res, err := a.client.ZeroTrust.Access.Groups.Update(ctx, ag.ID, params)
	if err != nil {
		return cftypes.AccessGroup{}, errors.Wrap(err, "unable to update access groups")
	}

	return parseAccessGroupFromRaw(
		res.ID,
		res.Name,
		res.Include,
		res.Exclude,
		res.Require,
		res.JSON.RawJSON(),
	), nil
}

func (a *API) DeleteAccessGroup(ctx context.Context, groupID string) error {
	_, err := a.client.ZeroTrust.Access.Groups.Delete(ctx, groupID, zero_trust.AccessGroupDeleteParams{
		AccountID: cloudflare.F(a.CFAccountID),
	})

	return errors.Wrap(err, "unable to delete access group")
}

func parseAccessApplicationFromRaw(
	id, name, domain string,
	appType zero_trust.ApplicationType,
	appLauncherVisible, autoRedirectToIdentity, enableBindingCookie, httpOnlyCookieAttribute bool,
	logoURL, sessionDuration, rawJSON string,
) cftypes.AccessApplication {
	createdAt, updatedAt := parseCreatedAndUpdated(rawJSON)

	allowedIdps := []string{}
	if idpsResult := gjson.Get(rawJSON, "allowed_idps"); idpsResult.IsArray() {
		for _, item := range idpsResult.Array() {
			if item.String() != "" {
				allowedIdps = append(allowedIdps, item.String())
			}
		}
	}

	policyIDs := []string{}
	if polsResult := gjson.Get(rawJSON, "policies"); polsResult.IsArray() {
		for _, item := range polsResult.Array() {
			if item.IsObject() {
				if pid := item.Get("id").String(); pid != "" && item.Get("reusable").Bool() {
					policyIDs = append(policyIDs, pid)
				}
			} else if item.String() != "" {
				policyIDs = append(policyIDs, item.String())
			}
		}
	}

	launcherVisible := appLauncherVisible
	autoRedirect := autoRedirectToIdentity
	bindingCookie := enableBindingCookie
	httpOnlyCookie := httpOnlyCookieAttribute

	return cftypes.AccessApplication{
		ID:                      id,
		Name:                    name,
		Domain:                  domain,
		Type:                    appType,
		AppLauncherVisible:      &launcherVisible,
		AllowedIdps:             allowedIdps,
		AutoRedirectToIdentity:  &autoRedirect,
		SessionDuration:         sessionDuration,
		EnableBindingCookie:     &bindingCookie,
		HttpOnlyCookieAttribute: &httpOnlyCookie,
		LogoURL:                 logoURL,
		CreatedAt:               createdAt,
		UpdatedAt:               updatedAt,
		PolicyIDs:               policyIDs,
	}
}

func (a *API) AccessApplications(ctx context.Context) ([]cftypes.AccessApplication, error) {
	pager := a.client.ZeroTrust.Access.Applications.ListAutoPaging(ctx, zero_trust.AccessApplicationListParams{
		AccountID: cloudflare.F(a.CFAccountID),
	})

	apps := []cftypes.AccessApplication{}
	for pager.Next() {
		item := pager.Current()
		apps = append(apps, parseAccessApplicationFromRaw(
			item.ID,
			item.Name,
			item.Domain,
			item.Type,
			item.AppLauncherVisible,
			item.AutoRedirectToIdentity,
			item.EnableBindingCookie,
			item.HTTPOnlyCookieAttribute,
			item.LogoURL,
			item.SessionDuration,
			item.JSON.RawJSON(),
		))
	}

	if err := pager.Err(); err != nil {
		return apps, errors.Wrap(err, "unable to get access applications")
	}

	return apps, nil
}

func (a *API) FindAccessApplicationByDomain(ctx context.Context, domain string) (*cftypes.AccessApplication, error) {
	apps, err := a.AccessApplications(ctx)
	if err != nil {
		return nil, err
	}

	for i := range apps {
		if apps[i].Domain == domain {
			return &apps[i], nil
		}
	}

	return nil, nil
}

func (a *API) AccessApplication(ctx context.Context, accessApplicationID string) (cftypes.AccessApplication, error) {
	res, err := a.client.ZeroTrust.Access.Applications.Get(ctx, accessApplicationID, zero_trust.AccessApplicationGetParams{
		AccountID: cloudflare.F(a.CFAccountID),
	})
	if err != nil {
		return cftypes.AccessApplication{}, errors.Wrap(err, "unable to get access application")
	}

	return parseAccessApplicationFromRaw(
		res.ID,
		res.Name,
		res.Domain,
		res.Type,
		res.AppLauncherVisible,
		res.AutoRedirectToIdentity,
		res.EnableBindingCookie,
		res.HTTPOnlyCookieAttribute,
		res.LogoURL,
		res.SessionDuration,
		res.JSON.RawJSON(),
	), nil
}

func buildNewAccessApplicationBody(ag cftypes.AccessApplication) zero_trust.AccessApplicationNewParamsBodyUnion {
	appType := ag.Type
	if appType == "" {
		appType = zero_trust.ApplicationTypeSelfHosted
	}

	if appType == zero_trust.ApplicationTypeSelfHosted {
		body := zero_trust.AccessApplicationNewParamsBodySelfHostedApplication{
			Domain: cloudflare.F(ag.Domain),
			Type:   cloudflare.F(appType),
			Name:   cloudflare.F(ag.Name),
		}
		if ag.AllowedIdps != nil {
			body.AllowedIdPs = cloudflare.F(ag.AllowedIdps)
		}
		if ag.AppLauncherVisible != nil {
			body.AppLauncherVisible = cloudflare.F(*ag.AppLauncherVisible)
		}
		if ag.AutoRedirectToIdentity != nil {
			body.AutoRedirectToIdentity = cloudflare.F(*ag.AutoRedirectToIdentity)
		}
		if ag.EnableBindingCookie != nil {
			body.EnableBindingCookie = cloudflare.F(*ag.EnableBindingCookie)
		}
		if ag.HttpOnlyCookieAttribute != nil {
			body.HTTPOnlyCookieAttribute = cloudflare.F(*ag.HttpOnlyCookieAttribute)
		}
		if ag.LogoURL != "" {
			body.LogoURL = cloudflare.F(ag.LogoURL)
		}
		if ag.SessionDuration != "" {
			body.SessionDuration = cloudflare.F(ag.SessionDuration)
		}

		return body
	}

	body := zero_trust.AccessApplicationNewParamsBody{
		Domain: cloudflare.F(ag.Domain),
		Type:   cloudflare.F(appType),
		Name:   cloudflare.F(ag.Name),
	}
	if ag.AllowedIdps != nil {
		body.AllowedIdPs = cloudflare.F[interface{}](ag.AllowedIdps)
	}
	if ag.AppLauncherVisible != nil {
		body.AppLauncherVisible = cloudflare.F(*ag.AppLauncherVisible)
	}
	if ag.AutoRedirectToIdentity != nil {
		body.AutoRedirectToIdentity = cloudflare.F(*ag.AutoRedirectToIdentity)
	}
	if ag.EnableBindingCookie != nil {
		body.EnableBindingCookie = cloudflare.F(*ag.EnableBindingCookie)
	}
	if ag.HttpOnlyCookieAttribute != nil {
		body.HTTPOnlyCookieAttribute = cloudflare.F(*ag.HttpOnlyCookieAttribute)
	}
	if ag.LogoURL != "" {
		body.LogoURL = cloudflare.F(ag.LogoURL)
	}
	if ag.SessionDuration != "" {
		body.SessionDuration = cloudflare.F(ag.SessionDuration)
	}

	return body
}

func buildUpdateAccessApplicationBody(ag cftypes.AccessApplication, setPolicies bool, policyIDs []string) zero_trust.AccessApplicationUpdateParamsBodyUnion {
	appType := ag.Type
	if appType == "" {
		appType = zero_trust.ApplicationTypeSelfHosted
	}

	if appType == zero_trust.ApplicationTypeSelfHosted {
		body := zero_trust.AccessApplicationUpdateParamsBodySelfHostedApplication{
			Domain: cloudflare.F(ag.Domain),
			Type:   cloudflare.F(appType),
			Name:   cloudflare.F(ag.Name),
		}
		if ag.AllowedIdps != nil {
			body.AllowedIdPs = cloudflare.F(ag.AllowedIdps)
		}
		if ag.AppLauncherVisible != nil {
			body.AppLauncherVisible = cloudflare.F(*ag.AppLauncherVisible)
		}
		if ag.AutoRedirectToIdentity != nil {
			body.AutoRedirectToIdentity = cloudflare.F(*ag.AutoRedirectToIdentity)
		}
		if ag.EnableBindingCookie != nil {
			body.EnableBindingCookie = cloudflare.F(*ag.EnableBindingCookie)
		}
		if ag.HttpOnlyCookieAttribute != nil {
			body.HTTPOnlyCookieAttribute = cloudflare.F(*ag.HttpOnlyCookieAttribute)
		}
		if ag.LogoURL != "" {
			body.LogoURL = cloudflare.F(ag.LogoURL)
		}
		if ag.SessionDuration != "" {
			body.SessionDuration = cloudflare.F(ag.SessionDuration)
		}
		if setPolicies {
			links := make([]zero_trust.AccessApplicationUpdateParamsBodySelfHostedApplicationPolicyUnion, 0, len(policyIDs))
			for i, id := range policyIDs {
				links = append(links, zero_trust.AccessApplicationUpdateParamsBodySelfHostedApplicationPoliciesAccessAppPolicyLink{
					ID:         cloudflare.F(id),
					Precedence: cloudflare.F(int64(i + 1)),
				})
			}
			body.Policies = cloudflare.F(links)
		}

		return body
	}

	body := zero_trust.AccessApplicationUpdateParamsBody{
		Domain: cloudflare.F(ag.Domain),
		Type:   cloudflare.F(appType),
		Name:   cloudflare.F(ag.Name),
	}
	if ag.AllowedIdps != nil {
		body.AllowedIdPs = cloudflare.F[interface{}](ag.AllowedIdps)
	}
	if ag.AppLauncherVisible != nil {
		body.AppLauncherVisible = cloudflare.F(*ag.AppLauncherVisible)
	}
	if ag.AutoRedirectToIdentity != nil {
		body.AutoRedirectToIdentity = cloudflare.F(*ag.AutoRedirectToIdentity)
	}
	if ag.EnableBindingCookie != nil {
		body.EnableBindingCookie = cloudflare.F(*ag.EnableBindingCookie)
	}
	if ag.HttpOnlyCookieAttribute != nil {
		body.HTTPOnlyCookieAttribute = cloudflare.F(*ag.HttpOnlyCookieAttribute)
	}
	if ag.LogoURL != "" {
		body.LogoURL = cloudflare.F(ag.LogoURL)
	}
	if ag.SessionDuration != "" {
		body.SessionDuration = cloudflare.F(ag.SessionDuration)
	}
	if setPolicies {
		type policyLink struct {
			ID         string `json:"id"`
			Precedence int64  `json:"precedence"`
		}
		links := make([]policyLink, 0, len(policyIDs))
		for i, id := range policyIDs {
			links = append(links, policyLink{ID: id, Precedence: int64(i + 1)})
		}
		body.Policies = cloudflare.F[interface{}](links)
	}

	return body
}

func (a *API) CreateAccessApplication(ctx context.Context, ag cftypes.AccessApplication) (cftypes.AccessApplication, error) {
	res, err := a.client.ZeroTrust.Access.Applications.New(ctx, zero_trust.AccessApplicationNewParams{
		AccountID: cloudflare.F(a.CFAccountID),
		Body:      buildNewAccessApplicationBody(ag),
	})
	if err != nil {
		return cftypes.AccessApplication{}, errors.Wrap(err, "unable to create access applications")
	}

	return parseAccessApplicationFromRaw(
		res.ID,
		res.Name,
		res.Domain,
		res.Type,
		res.AppLauncherVisible,
		res.AutoRedirectToIdentity,
		res.EnableBindingCookie,
		res.HTTPOnlyCookieAttribute,
		res.LogoURL,
		res.SessionDuration,
		res.JSON.RawJSON(),
	), nil
}

func (a *API) UpdateAccessApplication(ctx context.Context, ag cftypes.AccessApplication) (cftypes.AccessApplication, error) {
	setPolicies := len(ag.PolicyIDs) > 0
	res, err := a.client.ZeroTrust.Access.Applications.Update(ctx, ag.ID, zero_trust.AccessApplicationUpdateParams{
		AccountID: cloudflare.F(a.CFAccountID),
		Body:      buildUpdateAccessApplicationBody(ag, setPolicies, ag.PolicyIDs),
	})
	if err != nil {
		return cftypes.AccessApplication{}, errors.Wrap(err, "unable to update access applications")
	}

	return parseAccessApplicationFromRaw(
		res.ID,
		res.Name,
		res.Domain,
		res.Type,
		res.AppLauncherVisible,
		res.AutoRedirectToIdentity,
		res.EnableBindingCookie,
		res.HTTPOnlyCookieAttribute,
		res.LogoURL,
		res.SessionDuration,
		res.JSON.RawJSON(),
	), nil
}

func (a *API) SetAccessApplicationPolicies(ctx context.Context, ag cftypes.AccessApplication, policyIDs []string) (cftypes.AccessApplication, error) {
	res, err := a.client.ZeroTrust.Access.Applications.Update(ctx, ag.ID, zero_trust.AccessApplicationUpdateParams{
		AccountID: cloudflare.F(a.CFAccountID),
		Body:      buildUpdateAccessApplicationBody(ag, true, policyIDs),
	})
	if err != nil {
		return cftypes.AccessApplication{}, errors.Wrap(err, "unable to update access application policies")
	}

	return parseAccessApplicationFromRaw(
		res.ID,
		res.Name,
		res.Domain,
		res.Type,
		res.AppLauncherVisible,
		res.AutoRedirectToIdentity,
		res.EnableBindingCookie,
		res.HTTPOnlyCookieAttribute,
		res.LogoURL,
		res.SessionDuration,
		res.JSON.RawJSON(),
	), nil
}

func (a *API) DeleteAccessApplication(ctx context.Context, appID string) error {
	policies, _ := a.AccessPolicies(ctx, appID)

	_, err := a.client.ZeroTrust.Access.Applications.Delete(ctx, appID, zero_trust.AccessApplicationDeleteParams{
		AccountID: cloudflare.F(a.CFAccountID),
	})
	if err != nil {
		return errors.Wrap(err, "unable to delete access application")
	}

	for _, pol := range policies {
		if pol.Reusable && pol.ID != "" {
			polInfo, getErr := a.GetAccessPolicy(ctx, pol.ID)
			if getErr == nil && polInfo.AppCount == 0 {
				_ = a.DeleteAccessPolicy(ctx, pol.ID)
			}
		}
	}

	return nil
}

func (a *API) AccessPolicies(ctx context.Context, appID string) (cfcollections.AccessPolicyCollection, error) {
	pager := a.client.ZeroTrust.Access.Applications.Policies.ListAutoPaging(ctx, appID, zero_trust.AccessApplicationPolicyListParams{
		AccountID: cloudflare.F(a.CFAccountID),
	})

	policies := cfcollections.AccessPolicyCollection{}
	for pager.Next() {
		item := pager.Current()
		raw := item.JSON.RawJSON()
		createdAt, updatedAt := parseCreatedAndUpdated(raw)
		policies = append(policies, cftypes.AccessPolicy{
			ID:         item.ID,
			Name:       item.Name,
			Precedence: int(item.Precedence),
			Decision:   string(item.Decision),
			Reusable:   gjson.Get(raw, "reusable").Bool(),
			AppCount:   gjson.Get(raw, "app_count").Int(),
			Include:    parseRulesFromCloudflare(item.Include),
			Exclude:    parseRulesFromCloudflare(item.Exclude),
			Require:    parseRulesFromCloudflare(item.Require),
			CreatedAt:  createdAt,
			UpdatedAt:  updatedAt,
		})
	}

	if err := pager.Err(); err != nil {
		return policies, errors.Wrap(err, "unable to get access policies")
	}

	return policies, nil
}

func (a *API) GetAccessPolicy(ctx context.Context, policyID string) (cftypes.AccessPolicy, error) {
	res, err := a.client.ZeroTrust.Access.Policies.Get(ctx, policyID, zero_trust.AccessPolicyGetParams{
		AccountID: cloudflare.F(a.CFAccountID),
	})
	if err != nil {
		return cftypes.AccessPolicy{}, errors.Wrap(err, "unable to get access policy")
	}

	raw := res.JSON.RawJSON()
	createdAt, updatedAt := parseCreatedAndUpdated(raw)

	return cftypes.AccessPolicy{
		ID:        res.ID,
		Name:      res.Name,
		Decision:  string(res.Decision),
		Reusable:  bool(res.Reusable),
		AppCount:  res.AppCount,
		Include:   parseRulesFromCloudflare(res.Include),
		Exclude:   parseRulesFromCloudflare(res.Exclude),
		Require:   parseRulesFromCloudflare(res.Require),
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}, nil
}

func (a *API) CreateAccessPolicy(ctx context.Context, ag cftypes.AccessPolicy) (cftypes.AccessPolicy, error) {
	decision := zero_trust.Decision(ag.Decision)
	if decision == "" {
		decision = zero_trust.DecisionAllow
	}

	params := zero_trust.AccessPolicyNewParams{
		AccountID: cloudflare.F(a.CFAccountID),
		Name:      cloudflare.F(ag.Name),
		Decision:  cloudflare.F(decision),
	}
	if typed, ok := toAccessRuleUnionParams(ag.Include); ok {
		params.Include = cloudflare.F(typed)
	} else {
		params.Include = cloudflare.Raw[[]zero_trust.AccessRuleUnionParam](ag.Include)
	}
	if len(ag.Exclude) > 0 {
		if typed, ok := toAccessRuleUnionParams(ag.Exclude); ok {
			params.Exclude = cloudflare.F(typed)
		} else {
			params.Exclude = cloudflare.Raw[[]zero_trust.AccessRuleUnionParam](ag.Exclude)
		}
	}
	if len(ag.Require) > 0 {
		if typed, ok := toAccessRuleUnionParams(ag.Require); ok {
			params.Require = cloudflare.F(typed)
		} else {
			params.Require = cloudflare.Raw[[]zero_trust.AccessRuleUnionParam](ag.Require)
		}
	}
	if ag.IsolationRequired != nil {
		params.IsolationRequired = cloudflare.F(*ag.IsolationRequired)
	}
	if ag.SessionDuration != nil {
		params.SessionDuration = cloudflare.F(*ag.SessionDuration)
	}
	if ag.PurposeJustificationRequired != nil {
		params.PurposeJustificationRequired = cloudflare.F(*ag.PurposeJustificationRequired)
	}
	if ag.PurposeJustificationPrompt != nil {
		params.PurposeJustificationPrompt = cloudflare.F(*ag.PurposeJustificationPrompt)
	}
	if ag.ApprovalRequired != nil {
		params.ApprovalRequired = cloudflare.F(*ag.ApprovalRequired)
	}

	res, err := a.client.ZeroTrust.Access.Policies.New(ctx, params)
	if err != nil {
		return cftypes.AccessPolicy{}, errors.Wrap(err, "unable to create access policy")
	}

	raw := res.JSON.RawJSON()
	createdAt, updatedAt := parseCreatedAndUpdated(raw)

	return cftypes.AccessPolicy{
		ID:         res.ID,
		Name:       res.Name,
		Precedence: ag.Precedence,
		Decision:   string(res.Decision),
		Reusable:   bool(res.Reusable),
		AppCount:   res.AppCount,
		Include:    parseRulesFromCloudflare(res.Include),
		Exclude:    parseRulesFromCloudflare(res.Exclude),
		Require:    parseRulesFromCloudflare(res.Require),
		CreatedAt:  createdAt,
		UpdatedAt:  updatedAt,
	}, nil
}

func (a *API) UpdateAccessPolicy(ctx context.Context, ag cftypes.AccessPolicy) (cftypes.AccessPolicy, error) {
	decision := zero_trust.Decision(ag.Decision)
	if decision == "" {
		decision = zero_trust.DecisionAllow
	}

	params := zero_trust.AccessPolicyUpdateParams{
		AccountID: cloudflare.F(a.CFAccountID),
		Name:      cloudflare.F(ag.Name),
		Decision:  cloudflare.F(decision),
	}
	if typed, ok := toAccessRuleUnionParams(ag.Include); ok {
		params.Include = cloudflare.F(typed)
	} else {
		params.Include = cloudflare.Raw[[]zero_trust.AccessRuleUnionParam](ag.Include)
	}
	if typed, ok := toAccessRuleUnionParams(ag.Exclude); ok {
		params.Exclude = cloudflare.F(typed)
	} else {
		params.Exclude = cloudflare.Raw[[]zero_trust.AccessRuleUnionParam](ag.Exclude)
	}
	if typed, ok := toAccessRuleUnionParams(ag.Require); ok {
		params.Require = cloudflare.F(typed)
	} else {
		params.Require = cloudflare.Raw[[]zero_trust.AccessRuleUnionParam](ag.Require)
	}
	if ag.IsolationRequired != nil {
		params.IsolationRequired = cloudflare.F(*ag.IsolationRequired)
	}
	if ag.SessionDuration != nil {
		params.SessionDuration = cloudflare.F(*ag.SessionDuration)
	}
	if ag.PurposeJustificationRequired != nil {
		params.PurposeJustificationRequired = cloudflare.F(*ag.PurposeJustificationRequired)
	}
	if ag.PurposeJustificationPrompt != nil {
		params.PurposeJustificationPrompt = cloudflare.F(*ag.PurposeJustificationPrompt)
	}
	if ag.ApprovalRequired != nil {
		params.ApprovalRequired = cloudflare.F(*ag.ApprovalRequired)
	}

	res, err := a.client.ZeroTrust.Access.Policies.Update(ctx, ag.ID, params)
	if err != nil {
		return cftypes.AccessPolicy{}, errors.Wrap(err, "unable to update access policy")
	}

	raw := res.JSON.RawJSON()
	createdAt, updatedAt := parseCreatedAndUpdated(raw)

	return cftypes.AccessPolicy{
		ID:         res.ID,
		Name:       res.Name,
		Precedence: ag.Precedence,
		Decision:   string(res.Decision),
		Reusable:   bool(res.Reusable),
		AppCount:   res.AppCount,
		Include:    parseRulesFromCloudflare(res.Include),
		Exclude:    parseRulesFromCloudflare(res.Exclude),
		Require:    parseRulesFromCloudflare(res.Require),
		CreatedAt:  createdAt,
		UpdatedAt:  updatedAt,
	}, nil
}

func (a *API) DeleteAccessPolicy(ctx context.Context, policyID string) error {
	_, err := a.client.ZeroTrust.Access.Policies.Delete(ctx, policyID, zero_trust.AccessPolicyDeleteParams{
		AccountID: cloudflare.F(a.CFAccountID),
	})

	return errors.Wrap(err, "unable to delete access policy")
}

// MakeAccessPolicyReusable converts a legacy application-scoped policy into an account-level reusable policy.
func (a *API) MakeAccessPolicyReusable(ctx context.Context, appID string, policyID string) error {
	path := fmt.Sprintf("accounts/%s/access/apps/%s/policies/%s/make_reusable", a.CFAccountID, appID, policyID)
	err := a.client.Put(ctx, path, map[string]interface{}{}, nil)

	return errors.Wrap(err, "unable to convert legacy policy to reusable policy")
}

// CreateLegacyApplicationPolicy creates an application-scoped (legacy) policy directly on an application.
// Used in tests to verify migration of pre-existing legacy policies to reusable policies.
func (a *API) CreateLegacyApplicationPolicy(ctx context.Context, appID string, ag cftypes.AccessPolicy) (string, error) {
	path := fmt.Sprintf("accounts/%s/access/apps/%s/policies", a.CFAccountID, appID)
	decision := ag.Decision
	if decision == "" {
		decision = string(zero_trust.DecisionAllow)
	}
	payload := map[string]interface{}{
		"name":       ag.Name,
		"decision":   decision,
		"precedence": ag.Precedence,
		"include":    ag.Include,
	}
	if len(ag.Exclude) > 0 {
		payload["exclude"] = ag.Exclude
	}
	if len(ag.Require) > 0 {
		payload["require"] = ag.Require
	}

	var rawResp map[string]interface{}
	if err := a.client.Post(ctx, path, payload, &rawResp); err != nil {
		return "", errors.Wrap(err, "unable to create legacy application policy")
	}

	rawBytes, _ := json.Marshal(rawResp) //nolint:errchkjson
	id := gjson.GetBytes(rawBytes, "result.id").String()

	return id, nil
}

func (a *API) DeleteLegacyApplicationPolicy(ctx context.Context, appID string, policyID string) error {
	_, err := a.client.ZeroTrust.Access.Applications.Policies.Delete(ctx, appID, policyID, zero_trust.AccessApplicationPolicyDeleteParams{
		AccountID: cloudflare.F(a.CFAccountID),
	})

	return errors.Wrap(err, "unable to delete legacy application policy")
}

func parseServiceToken(id, name, clientID, duration string, expiresAt time.Time, rawJSON string) cftypes.AccessServiceToken {
	createdAt, updatedAt := parseCreatedAndUpdated(rawJSON)
	exp := expiresAt
	if exp.IsZero() {
		if parsed := parseTimeField(rawJSON, "expires_at"); parsed != nil {
			exp = *parsed
		} else if createdAt != nil {
			if dur, err := time.ParseDuration(duration); err == nil && dur > 0 {
				exp = createdAt.Add(dur)
			} else {
				exp = createdAt.Add(8760 * time.Hour)
			}
		}
	}

	return cftypes.AccessServiceToken{
		ID:        id,
		Name:      name,
		ClientID:  clientID,
		Duration:  duration,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
		ExpiresAt: &exp,
	}
}

func (a *API) ServiceTokens(ctx context.Context) ([]cftypes.ExtendedServiceToken, error) {
	pager := a.client.ZeroTrust.Access.ServiceTokens.ListAutoPaging(ctx, zero_trust.AccessServiceTokenListParams{
		AccountID: cloudflare.F(a.CFAccountID),
	})

	extendedTokens := []cftypes.ExtendedServiceToken{}
	for pager.Next() {
		token := pager.Current()
		extendedTokens = append(extendedTokens, cftypes.ExtendedServiceToken{
			AccessServiceToken: parseServiceToken(
				token.ID,
				token.Name,
				token.ClientID,
				token.Duration,
				token.ExpiresAt,
				token.JSON.RawJSON(),
			),
		})
	}

	if err := pager.Err(); err != nil {
		return extendedTokens, errors.Wrap(err, "unable to get service tokens")
	}

	return extendedTokens, nil
}

func (a *API) CreateAccessServiceToken(ctx context.Context, token cftypes.ExtendedServiceToken) (cftypes.ExtendedServiceToken, error) {
	res, err := a.client.ZeroTrust.Access.ServiceTokens.New(ctx, zero_trust.AccessServiceTokenNewParams{
		AccountID: cloudflare.F(a.CFAccountID),
		Name:      cloudflare.F(token.Name),
	})
	if err != nil {
		return cftypes.ExtendedServiceToken{}, errors.Wrap(err, "unable to create access service token")
	}

	parsed := parseServiceToken(
		res.ID,
		res.Name,
		res.ClientID,
		res.Duration,
		time.Time{},
		res.JSON.RawJSON(),
	)

	return cftypes.ExtendedServiceToken{
		ClientSecret:       res.ClientSecret,
		AccessServiceToken: parsed,
	}, nil
}

func (a *API) UpdateAccessServiceToken(ctx context.Context, token cftypes.ExtendedServiceToken) (cftypes.ExtendedServiceToken, error) {
	res, err := a.client.ZeroTrust.Access.ServiceTokens.Update(ctx, token.ID, zero_trust.AccessServiceTokenUpdateParams{
		AccountID: cloudflare.F(a.CFAccountID),
		Name:      cloudflare.F(token.Name),
	})
	if err != nil {
		return token, errors.Wrap(err, "unable to update access service token")
	}

	parsed := parseServiceToken(
		res.ID,
		res.Name,
		res.ClientID,
		res.Duration,
		res.ExpiresAt,
		res.JSON.RawJSON(),
	)
	token.AccessServiceToken = parsed

	return token, nil
}

func (a *API) RotateAccessServiceToken(ctx context.Context, token cftypes.ExtendedServiceToken) (cftypes.ExtendedServiceToken, error) {
	res, err := a.client.ZeroTrust.Access.ServiceTokens.Rotate(ctx, token.ID, zero_trust.AccessServiceTokenRotateParams{
		AccountID: cloudflare.F(a.CFAccountID),
	})
	if err != nil {
		return cftypes.ExtendedServiceToken{}, errors.Wrap(err, "unable to rotate access service token")
	}

	parsed := parseServiceToken(
		res.ID,
		res.Name,
		res.ClientID,
		res.Duration,
		time.Time{},
		res.JSON.RawJSON(),
	)

	return cftypes.ExtendedServiceToken{
		ClientSecret:       res.ClientSecret,
		AccessServiceToken: parsed,
	}, nil
}

func (a *API) DeleteAccessServiceToken(ctx context.Context, tokenID string) error {
	_, err := a.client.ZeroTrust.Access.ServiceTokens.Delete(ctx, tokenID, zero_trust.AccessServiceTokenDeleteParams{
		AccountID: cloudflare.F(a.CFAccountID),
	})

	return errors.Wrap(err, "unable to delete access service token")
}
