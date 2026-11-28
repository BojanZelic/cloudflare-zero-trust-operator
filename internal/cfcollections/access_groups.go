package cfcollections

import (
	"encoding/json"
	"reflect"

	"github.com/bojanzelic/cloudflare-zero-trust-operator/internal/cftypes"
)

type AccessGroupCollection []cftypes.AccessGroup

func (c AccessGroupCollection) Len() int { return len(c) }

func (c AccessGroupCollection) GetByName(name string) *cftypes.AccessGroup {
	for i := range c {
		if c[i].Name == name {
			return &c[i]
		}
	}

	return nil
}

func AccessGroupEqual(first cftypes.AccessGroup, second cftypes.AccessGroup) bool {
	if first.Name != second.Name {
		return false
	}

	return ruleSlicesEqual(first.Include, second.Include) &&
		ruleSlicesEqual(first.Exclude, second.Exclude) &&
		ruleSlicesEqual(first.Require, second.Require)
}

func ruleSlicesEqual(first []interface{}, second []interface{}) bool {
	if len(first) == 0 && len(second) == 0 {
		return true
	}

	if len(first) != len(second) {
		return false
	}

	v1, _ := json.Marshal(first)  //nolint:errchkjson
	v2, _ := json.Marshal(second) //nolint:errchkjson

	return reflect.DeepEqual(v1, v2)
}
