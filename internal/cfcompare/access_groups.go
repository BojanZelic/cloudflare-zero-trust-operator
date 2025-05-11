package cfcompare

import (
	"encoding/json"
	"reflect"

	"github.com/bojanzelic/cloudflare-zero-trust-operator/api/v4alpha1"
	"github.com/cloudflare/cloudflare-go/v4/zero_trust"
)

//nolint:varnamelen
func AreAccessGroupsEquivalent(cf *zero_trust.AccessGroupGetResponse, k8s *v4alpha1.CloudflareAccessGroup) bool {
	if cf.Name != k8s.Spec.Name {
		return false
	}

	v1, _ := json.Marshal(cf.Include)       //nolint:errchkjson,varnamelen
	v2, _ := json.Marshal(k8s.Spec.Include) //nolint:errchkjson,varnamelen

	if !reflect.DeepEqual(v1, v2) {
		return false
	}

	v1, _ = json.Marshal(cf.Exclude)       //nolint:errchkjson
	v2, _ = json.Marshal(k8s.Spec.Exclude) //nolint:errchkjson

	if !reflect.DeepEqual(v1, v2) {
		return false
	}

	v1, _ = json.Marshal(cf.Require)       //nolint:errchkjson
	v2, _ = json.Marshal(k8s.Spec.Require) //nolint:errchkjson

	return reflect.DeepEqual(v1, v2)
}
