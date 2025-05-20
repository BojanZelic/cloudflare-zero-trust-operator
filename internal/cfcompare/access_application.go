package cfcompare

import (
	"context"
	"reflect"
	"strings"

	"github.com/Southclaws/fault"
	"github.com/Southclaws/fault/fctx"
	"github.com/Southclaws/fault/fmsg"
	"github.com/bojanzelic/cloudflare-zero-trust-operator/api/v4alpha1"
	"github.com/cloudflare/cloudflare-go/v4/zero_trust"
	"github.com/go-logr/logr"
)

//nolint:cyclop,varnamelen
func AreAccessApplicationsEquivalent(
	ctx context.Context,
	log *logr.Logger,
	cf *zero_trust.AccessApplicationGetResponse,
	k8s *v4alpha1.CloudflareAccessApplication,
) bool {
	//
	k8sType := strings.TrimSpace(k8s.Spec.Type)
	strsEq := strings.TrimSpace(cf.Type) == k8sType &&
		strings.TrimSpace(cf.SessionDuration) == strings.TrimSpace(k8s.Spec.SessionDuration) &&
		strings.TrimSpace(cf.LogoURL) == strings.TrimSpace(k8s.Spec.LogoURL)

	//
	boolsEq := (k8s.Spec.AppLauncherVisible == nil || reflect.DeepEqual(cf.AppLauncherVisible, &k8s.Spec.AppLauncherVisible)) &&
		(k8s.Spec.AutoRedirectToIdentity == nil || reflect.DeepEqual(cf.AutoRedirectToIdentity, &k8s.Spec.AutoRedirectToIdentity)) &&
		(k8s.Spec.EnableBindingCookie == nil || reflect.DeepEqual(cf.EnableBindingCookie, &k8s.Spec.EnableBindingCookie)) &&
		(k8s.Spec.HTTPOnlyCookieAttribute == nil || reflect.DeepEqual(cf.HTTPOnlyCookieAttribute, &k8s.Spec.HTTPOnlyCookieAttribute))

	//
	idpEq := (cf.AllowedIdPs == nil && len(k8s.Spec.AllowedIdps) == 0) || reflect.DeepEqual(cf.AllowedIdPs, k8s.Spec.AllowedIdps)

	//
	universal := DoCFPoliciesEquateToK8Ss(ctx, log, cf, k8s) &&
		boolsEq &&
		strsEq &&
		idpEq

	// specific equalities for app types
	switch k8sType {
	case string(zero_trust.ApplicationTypeSelfHosted):
		{
			// also check name and domain
			return universal &&
				strings.TrimSpace(cf.Name) == strings.TrimSpace(k8s.Spec.Name) &&
				strings.TrimSpace(cf.Domain) == strings.TrimSpace(k8s.Spec.Domain)
		}
	case string(zero_trust.ApplicationTypeAppLauncher),
		string(zero_trust.ApplicationTypeWARP):
		{
			return universal
		}
	}

	//
	// Explicitly unhandled
	//

	log.Error(
		fault.New("Access application comparaison for type is not explicitly handled",
			fmsg.With("Undetermined behavior of this operator is to be expected"),
			fctx.With(ctx,
				"appType", k8sType,
				"advice", "Contact the developers for a feature push",
			),
		),
		"This application type is most likely not handled by this operator yet.",
	)

	//
	return universal
}
