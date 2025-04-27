package cfcollections

import (
	"encoding/json"
	"reflect"
	"sort"

	"github.com/cloudflare/cloudflare-go/v4/zero_trust"
)

type LegacyAccessPolicyCollection []zero_trust.AccessApplicationPolicyListResponse

func (c LegacyAccessPolicyCollection) Len() int { return len(c) }

func (c LegacyAccessPolicyCollection) SortByPrecedence() {
	sort.Slice(c, func(i, j int) bool {
		return c[i].Precedence < c[j].Precedence
	})
}

//nolint:cyclop
func AccessPoliciesEqual(first *zero_trust.AccessApplicationPolicyListResponse, second *zero_trust.AccessApplicationPolicyListResponse) bool {
	if first == nil && second == nil {
		return true
	}

	if first == nil || second == nil {
		return false
	}

	if first.Name != second.Name {
		return false
	}
	if first.Precedence != second.Precedence {
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
