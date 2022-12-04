package cfcollections

import (
	"reflect"
	"strings"

	cloudflare "github.com/cloudflare/cloudflare-go"
)

func AccessAppEqual(first cloudflare.AccessApplication, second cloudflare.AccessApplication) bool {
	return strings.TrimSpace(first.Name) == strings.TrimSpace(second.Name) &&
		strings.TrimSpace(first.Domain) == strings.TrimSpace(second.Domain) &&
		first.Type == second.Type &&
		reflect.DeepEqual(first.AppLauncherVisible, second.AppLauncherVisible) &&
		reflect.DeepEqual(first.AutoRedirectToIdentity, second.AutoRedirectToIdentity) &&
		reflect.DeepEqual(first.EnableBindingCookie, second.EnableBindingCookie) &&
		reflect.DeepEqual(first.HttpOnlyCookieAttribute, second.HttpOnlyCookieAttribute) &&
		strings.TrimSpace(first.SessionDuration) == strings.TrimSpace(second.SessionDuration) &&
		reflect.DeepEqual(first.AllowedIdps, second.AllowedIdps)
}
