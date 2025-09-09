package cfapi

//
// These API call are most probably used only for tests
//

import (
	"slices"

	"github.com/Southclaws/fault"
	"github.com/Southclaws/fault/fmsg"
	"github.com/cloudflare/cloudflare-go/v4/shared"
	"github.com/cloudflare/cloudflare-go/v4/zero_trust"
)

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

//
//
//

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

//
//
//
//
//
//

// Extract a simple, precedence ordered slice of associated [app] policy UUIDs
func GetOrderedPolicyUUIDs(app *zero_trust.AccessApplicationGetResponse) (orderedUUIDs []string, err error) {
	switch v := app.Policies.(type) { //nolint:varnamelen
	case []zero_trust.AccessApplicationGetResponseDeviceEnrollmentPermissionsApplicationPolicy:
		{
			return getOrderedPolicyUUIDs_WARP(v), nil
		}
	case []zero_trust.AccessApplicationGetResponseAppLauncherApplicationPolicy:
		{
			return getOrderedPolicyUUIDs_AppLauncher(v), nil
		}
	case []zero_trust.AccessApplicationGetResponseSelfHostedApplicationPolicy:
		{
			// @dev should be needed, but as of cfgo-v' 4.4.0, generic AccessApplicationGetResponse produces policies of
			// [[]zero_trust.AccessApplicationGetResponseSelfHostedApplicationPolicy], whatever the type of app
			return getOrderedPolicyUUIDs_SelfHosted(v), nil
		}
	}

	return nil, fault.New("Cannot retrieve policy UUIDs", fmsg.With("Unhandled policy, contact the developers")) //nolint:wrapcheck
}

//
//
//

func getOrderedPolicyUUIDs_SelfHosted(policies []zero_trust.AccessApplicationGetResponseSelfHostedApplicationPolicy) (orderedUUIDs []string) {
	slices.SortStableFunc(policies, func(a, b zero_trust.AccessApplicationGetResponseSelfHostedApplicationPolicy) int {
		if a.Precedence < b.Precedence {
			return -1
		} else if a.Precedence > b.Precedence {
			return 1
		}
		return 0
	})
	for _, orderedPolicies := range policies {
		orderedUUIDs = append(orderedUUIDs, orderedPolicies.ID)
	}
	return orderedUUIDs
}

func getOrderedPolicyUUIDs_WARP(policies []zero_trust.AccessApplicationGetResponseDeviceEnrollmentPermissionsApplicationPolicy) (orderedUUIDs []string) {
	slices.SortStableFunc(policies, func(a, b zero_trust.AccessApplicationGetResponseDeviceEnrollmentPermissionsApplicationPolicy) int {
		if a.Precedence < b.Precedence {
			return -1
		} else if a.Precedence > b.Precedence {
			return 1
		}
		return 0
	})
	for _, orderedPolicies := range policies {
		orderedUUIDs = append(orderedUUIDs, orderedPolicies.ID)
	}
	return orderedUUIDs
}

func getOrderedPolicyUUIDs_AppLauncher(policies []zero_trust.AccessApplicationGetResponseAppLauncherApplicationPolicy) (orderedUUIDs []string) {
	slices.SortStableFunc(policies, func(a, b zero_trust.AccessApplicationGetResponseAppLauncherApplicationPolicy) int {
		if a.Precedence < b.Precedence {
			return -1
		} else if a.Precedence > b.Precedence {
			return 1
		}
		return 0
	})
	for _, orderedPolicies := range policies {
		orderedUUIDs = append(orderedUUIDs, orderedPolicies.ID)
	}
	return orderedUUIDs
}

//
//
//

//nolint:varnamelen
func difference(a, b []string) []string {
	m := make(map[string]struct{}, len(b))
	for _, s := range b {
		m[s] = struct{}{}
	}

	var diff []string
	for _, s := range a {
		if _, found := m[s]; !found {
			diff = append(diff, s)
		}
	}
	return diff
}
