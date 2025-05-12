package cfcompare

import (
	"reflect"
	"strings"

	"github.com/bojanzelic/cloudflare-zero-trust-operator/api/v4alpha1"
	"github.com/cloudflare/cloudflare-go/v4/zero_trust"
)

// TODO fix unit test

//nolint:cyclop,varnamelen
func AreAccessApplicationsEquivalent(cf *zero_trust.AccessApplicationGetResponse, k8s *v4alpha1.CloudflareAccessApplication) bool {
	//
	k8sType := strings.TrimSpace(k8s.Spec.Type)
	if k8sType == "" {
		k8sType = string(zero_trust.ApplicationTypeSelfHosted)
	}

	//
	universal := DoK8SAccessPoliciesMatch(cf, k8s) &&
		strings.TrimSpace(cf.Type) == strings.TrimSpace(k8sType) &&
		strings.TrimSpace(cf.SessionDuration) == strings.TrimSpace(k8s.Spec.SessionDuration) &&
		reflect.DeepEqual(cf.AppLauncherVisible, k8s.Spec.AppLauncherVisible) &&
		reflect.DeepEqual(cf.AutoRedirectToIdentity, k8s.Spec.AutoRedirectToIdentity) &&
		reflect.DeepEqual(cf.EnableBindingCookie, k8s.Spec.EnableBindingCookie) &&
		reflect.DeepEqual(cf.HTTPOnlyCookieAttribute, k8s.Spec.HTTPOnlyCookieAttribute) &&
		reflect.DeepEqual(cf.LogoURL, k8s.Spec.LogoURL) &&
		reflect.DeepEqual(cf.AllowedIdPs, k8s.Spec.AllowedIdps)

	//
	switch k8sType {
	case string(zero_trust.ApplicationTypeSelfHosted):
		{
			// also check name and domain
			return universal &&
				strings.TrimSpace(cf.Name) == strings.TrimSpace(k8s.Spec.Name) &&
				strings.TrimSpace(cf.Domain) == strings.TrimSpace(k8s.Spec.Domain)
		}
	case string(zero_trust.ApplicationTypeWARP):
	case string(zero_trust.ApplicationTypeAppLauncher):
	default:
		{
			// universal behavior
		}
	}

	return universal
}
