package cfcollections

import (
	"reflect"
	"strings"

	"github.com/bojanzelic/cloudflare-zero-trust-operator/internal/cftypes"
)

func AccessAppEqual(first cftypes.AccessApplication, second cftypes.AccessApplication) bool {
	allowedIdpsEqual := (len(first.AllowedIdps) == 0 && len(second.AllowedIdps) == 0) ||
		reflect.DeepEqual(first.AllowedIdps, second.AllowedIdps)

	return strings.TrimSpace(first.Name) == strings.TrimSpace(second.Name) &&
		strings.TrimSpace(first.Domain) == strings.TrimSpace(second.Domain) &&
		first.Type == second.Type &&
		reflect.DeepEqual(first.AppLauncherVisible, second.AppLauncherVisible) &&
		reflect.DeepEqual(first.AutoRedirectToIdentity, second.AutoRedirectToIdentity) &&
		reflect.DeepEqual(first.EnableBindingCookie, second.EnableBindingCookie) &&
		reflect.DeepEqual(first.HttpOnlyCookieAttribute, second.HttpOnlyCookieAttribute) &&
		reflect.DeepEqual(first.LogoURL, second.LogoURL) &&
		strings.TrimSpace(first.SessionDuration) == strings.TrimSpace(second.SessionDuration) &&
		allowedIdpsEqual
}
