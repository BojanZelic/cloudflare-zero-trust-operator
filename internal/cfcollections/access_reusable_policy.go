package cfcollections

import (
	"encoding/json"
	"reflect"

	"github.com/bojanzelic/cloudflare-zero-trust-operator/api/v4alpha1"
	"github.com/cloudflare/cloudflare-go/v4/zero_trust"
)

type AccessReusablePolicyCollection []zero_trust.AccessPolicyListResponse

func (c AccessReusablePolicyCollection) Len() int { return len(c) }

//nolint:cyclop
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

	if len(cf.Include) != 0 && len(k8s.Spec.Include) != 0 {
		v1, _ := json.Marshal(cf.Include)       //nolint:errchkjson
		v2, _ := json.Marshal(k8s.Spec.Include) //nolint:errchkjson
		if !reflect.DeepEqual(v1, v2) {
			return false
		}
	}

	if len(cf.Exclude) != 0 && len(k8s.Spec.Exclude) != 0 {
		v1, _ := json.Marshal(cf.Exclude)       //nolint:errchkjson
		v2, _ := json.Marshal(k8s.Spec.Exclude) //nolint:errchkjson

		if !reflect.DeepEqual(v1, v2) {
			return false
		}
	}

	if len(cf.Require) != 0 && len(k8s.Spec.Require) != 0 {
		v1, _ := json.Marshal(cf.Require)       //nolint:errchkjson
		v2, _ := json.Marshal(k8s.Spec.Require) //nolint:errchkjson
		if !reflect.DeepEqual(v1, v2) {
			return false
		}
	}

	return true
}
