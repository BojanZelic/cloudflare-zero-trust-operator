package cfcompare

import (
	"encoding/json"
	"reflect"

	"github.com/bojanzelic/cloudflare-zero-trust-operator/api/v4alpha1"
	"github.com/cloudflare/cloudflare-go/v4/zero_trust"
)

// areAccessRulesEquivalent compare les règles d'accès CloudFlare avec celles de Kubernetes
func areAccessRulesEquivalent(cfRules, k8sRules []zero_trust.AccessRule) bool {
	if len(cfRules) == 0 || len(k8sRules) == 0 {
		return true // Si l'une des listes est vide, considérons qu'elles sont équivalentes
	}

	v1, _ := json.Marshal(cfRules)  //nolint:errchkjson
	v2, _ := json.Marshal(k8sRules) //nolint:errchkjson

	return reflect.DeepEqual(v1, v2)
}

//nolint:cyclop,varnamelen
func AreAccessReusablePoliciesEquivalent(cf *zero_trust.AccessPolicyGetResponse, k8s *v4alpha1.CloudflareAccessReusablePolicy) bool {
	if cf == nil && k8s == nil {
		return true
	}

	if cf == nil || k8s == nil {
		return false
	}

	if cf.Name != k8s.Spec.Name {
		return false
	}

	if !areAccessRulesEquivalent(cf.Include, k8s.Spec.Include.ToAccessRules(k8s.Status.ResolvedIdpsFromRefs.Include)) {
		return false
	}
	if !areAccessRulesEquivalent(cf.Exclude, k8s.Spec.Exclude.ToAccessRules(k8s.Status.ResolvedIdpsFromRefs.Exclude)) {
		return false
	}
	if !areAccessRulesEquivalent(cf.Require, k8s.Spec.Require.ToAccessRules(k8s.Status.ResolvedIdpsFromRefs.Require)) {
		return false
	}

	return true
}
