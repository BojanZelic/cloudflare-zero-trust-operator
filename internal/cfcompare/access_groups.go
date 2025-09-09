package cfcompare

import (
	"github.com/bojanzelic/cloudflare-zero-trust-operator/api/v4alpha1"
	"github.com/cloudflare/cloudflare-go/v4/zero_trust"
)

//nolint:cyclop,varnamelen
func AreAccessGroupsEquivalent(cf *zero_trust.AccessGroupGetResponse, k8s *v4alpha1.CloudflareAccessGroup) bool {
	if cf == nil && k8s == nil {
		return true
	}

	if cf == nil || k8s == nil {
		return false
	}

	if cf.Name != k8s.Spec.Name {
		return false
	}

	if !AreAccessRulesEquivalent(cf.Include, k8s.Spec.Include.ToAccessRules(k8s.Status.ResolvedIdpsFromRefs.Include)) {
		return false
	}
	if !AreAccessRulesEquivalent(cf.Exclude, k8s.Spec.Exclude.ToAccessRules(k8s.Status.ResolvedIdpsFromRefs.Exclude)) {
		return false
	}
	if !AreAccessRulesEquivalent(cf.Require, k8s.Spec.Require.ToAccessRules(k8s.Status.ResolvedIdpsFromRefs.Require)) {
		return false
	}

	return true
}
