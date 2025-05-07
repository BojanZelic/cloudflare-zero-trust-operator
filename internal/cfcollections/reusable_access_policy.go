package cfcollections

import (
	"encoding/json"
	"reflect"

	"github.com/cloudflare/cloudflare-go/v4/zero_trust"
)

type AccessReusablePolicyCollection []zero_trust.AccessPolicyListResponse

func (c AccessReusablePolicyCollection) Len() int { return len(c) }

//nolint:cyclop
func AccessReusablePoliciesEqual(first *zero_trust.AccessPolicyListResponse, second *zero_trust.AccessPolicyListResponse) bool {
	if first == nil && second == nil {
		return true
	}

	if first == nil || second == nil {
		return false
	}

	if first.Name != second.Name {
		return false
	}

	if len(first.Include) != 0 && len(second.Include) != 0 {
		v1, _ := json.Marshal(first.Include)  //nolint:errchkjson
		v2, _ := json.Marshal(second.Include) //nolint:errchkjson
		if !reflect.DeepEqual(v1, v2) {
			return false
		}
	}

	if len(first.Exclude) != 0 && len(second.Exclude) != 0 {
		v1, _ := json.Marshal(first.Exclude)  //nolint:errchkjson
		v2, _ := json.Marshal(second.Exclude) //nolint:errchkjson

		if !reflect.DeepEqual(v1, v2) {
			return false
		}
	}

	if len(first.Require) != 0 && len(second.Require) != 0 {
		v1, _ := json.Marshal(first.Require)  //nolint:errchkjson
		v2, _ := json.Marshal(second.Require) //nolint:errchkjson
		if !reflect.DeepEqual(v1, v2) {
			return false
		}
	}

	return true
}
