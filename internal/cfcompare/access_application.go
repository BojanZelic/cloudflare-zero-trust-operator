package cfcompare

import (
	"reflect"
	"strings"

	"github.com/bojanzelic/cloudflare-zero-trust-operator/api/v4alpha1"
	"github.com/cloudflare/cloudflare-go/v4/zero_trust"
)

//nolint:cyclop,varnamelen
func AreAccessApplicationsEquivalent(cf *zero_trust.AccessApplicationGetResponse, k8s *v4alpha1.CloudflareAccessApplication) bool {
	return DoK8SAccessPoliciesMatch(cf, k8s) &&
		strings.TrimSpace(cf.Type) == strings.TrimSpace(k8s.Spec.Type) &&
		strings.TrimSpace(cf.Name) == strings.TrimSpace(k8s.Spec.Name) &&
		strings.TrimSpace(cf.Domain) == strings.TrimSpace(k8s.Spec.Domain) &&
		strings.TrimSpace(cf.SessionDuration) == strings.TrimSpace(k8s.Spec.SessionDuration) &&
		reflect.DeepEqual(cf.AppLauncherVisible, k8s.Spec.AppLauncherVisible) &&
		reflect.DeepEqual(cf.AutoRedirectToIdentity, k8s.Spec.AutoRedirectToIdentity) &&
		reflect.DeepEqual(cf.EnableBindingCookie, k8s.Spec.EnableBindingCookie) &&
		reflect.DeepEqual(cf.HTTPOnlyCookieAttribute, k8s.Spec.HTTPOnlyCookieAttribute) &&
		reflect.DeepEqual(cf.LogoURL, k8s.Spec.LogoURL) &&
		reflect.DeepEqual(cf.AllowedIdPs, k8s.Spec.AllowedIdps)
}
