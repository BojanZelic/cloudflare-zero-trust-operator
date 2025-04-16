package cfcollections

import (
	"encoding/json"
	"reflect"

	"github.com/cloudflare/cloudflare-go/v4/zero_trust"
)

type AccessGroupCollection []zero_trust.AccessGroupListResponse

func (c AccessGroupCollection) Len() int { return len(c) }

func (c AccessGroupCollection) GetByName(name string) *zero_trust.AccessGroupListResponse {
	for _, policy := range c {
		if policy.Name == name {
			return &policy
		}
	}

	return nil
}

func AccessGroupEqual(first zero_trust.AccessGroupListResponse, second zero_trust.AccessGroupListResponse) bool {
	if first.Name != second.Name {
		return false
	}

	v1, _ := json.Marshal(first.Include)  //nolint:errchkjson,varnamelen
	v2, _ := json.Marshal(second.Include) //nolint:errchkjson,varnamelen

	if !reflect.DeepEqual(v1, v2) {
		return false
	}

	v1, _ = json.Marshal(first.Exclude)  //nolint:errchkjson
	v2, _ = json.Marshal(second.Exclude) //nolint:errchkjson

	if !reflect.DeepEqual(v1, v2) {
		return false
	}

	v1, _ = json.Marshal(first.Require)  //nolint:errchkjson
	v2, _ = json.Marshal(second.Require) //nolint:errchkjson

	return reflect.DeepEqual(v1, v2)
}
