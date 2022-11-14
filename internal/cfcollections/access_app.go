package cfcollections

import (
	"reflect"

	cloudflare "github.com/cloudflare/cloudflare-go"
)

func AccessAppEqual(first cloudflare.AccessApplication, second cloudflare.AccessApplication) bool {
	return first.Name == second.Name &&
		first.Domain == second.Domain &&
		first.Type == second.Type &&
		reflect.DeepEqual(first.AppLauncherVisible, second.AppLauncherVisible) &&
		reflect.DeepEqual(first.AutoRedirectToIdentity, second.AutoRedirectToIdentity) &&
		reflect.DeepEqual(first.EnableBindingCookie, second.EnableBindingCookie) &&
		reflect.DeepEqual(first.HttpOnlyCookieAttribute, second.HttpOnlyCookieAttribute) &&
		first.SessionDuration == second.SessionDuration &&
		reflect.DeepEqual(first.AllowedIdps, second.AllowedIdps)
}
