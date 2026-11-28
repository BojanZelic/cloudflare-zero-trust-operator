package cfcollections

import (
	"sort"

	"github.com/bojanzelic/cloudflare-zero-trust-operator/internal/cftypes"
)

type AccessPolicyCollection []cftypes.AccessPolicy

func (c AccessPolicyCollection) Len() int { return len(c) }

func (c AccessPolicyCollection) SortByPrecidence() {
	sort.Slice(c, func(i, j int) bool {
		return c[i].Precedence < c[j].Precedence
	})
}

func AccessPolicyContentEqual(first *cftypes.AccessPolicy, second *cftypes.AccessPolicy) bool {
	if first == nil && second == nil {
		return true
	}

	if first == nil || second == nil {
		return false
	}

	if first.Name != second.Name {
		return false
	}

	if first.Decision != "" && second.Decision != "" && first.Decision != second.Decision {
		return false
	}

	return ruleSlicesEqual(first.Include, second.Include) &&
		ruleSlicesEqual(first.Exclude, second.Exclude) &&
		ruleSlicesEqual(first.Require, second.Require)
}

func AccessPoliciesEqual(first *cftypes.AccessPolicy, second *cftypes.AccessPolicy) bool {
	if first == nil && second == nil {
		return true
	}

	if first == nil || second == nil {
		return false
	}

	if first.Precedence != second.Precedence {
		return false
	}

	return AccessPolicyContentEqual(first, second)
}
