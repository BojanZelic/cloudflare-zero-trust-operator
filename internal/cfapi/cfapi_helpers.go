package cfapi

//
// These API call are most probably used only for tests
//

import (
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
