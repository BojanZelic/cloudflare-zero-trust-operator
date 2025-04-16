package cfcollections

import (
	"reflect"
	"strings"

	"github.com/cloudflare/cloudflare-go/v4/zero_trust"
)

func AccessAppEqual(first zero_trust.AccessApplicationGetResponse, second zero_trust.AccessApplicationGetResponse) bool {
	return strings.TrimSpace(first.Name) == strings.TrimSpace(second.Name) &&
		strings.TrimSpace(first.Domain) == strings.TrimSpace(second.Domain) &&
		first.Type == second.Type &&
		reflect.DeepEqual(first.AppLauncherVisible, second.AppLauncherVisible) &&
		reflect.DeepEqual(first.AutoRedirectToIdentity, second.AutoRedirectToIdentity) &&
		reflect.DeepEqual(first.EnableBindingCookie, second.EnableBindingCookie) &&
		reflect.DeepEqual(first.HTTPOnlyCookieAttribute, second.HTTPOnlyCookieAttribute) &&
		reflect.DeepEqual(first.LogoURL, second.LogoURL) &&
		strings.TrimSpace(first.SessionDuration) == strings.TrimSpace(second.SessionDuration) &&
		reflect.DeepEqual(first.AllowedIdPs, second.AllowedIdPs)
}
