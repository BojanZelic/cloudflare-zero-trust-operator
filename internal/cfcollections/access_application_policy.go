package cfcollections

import (
	"sort"

	"github.com/bojanzelic/cloudflare-zero-trust-operator/api/v4alpha1"
	"github.com/cloudflare/cloudflare-go/v4/zero_trust"
)

type AccessApplicationPolicyCollection []zero_trust.AccessApplicationPolicyListResponse

func (c AccessApplicationPolicyCollection) Len() int { return len(c) }

func (c AccessApplicationPolicyCollection) SortByPrecedence() {
	sort.Slice(c, func(i, j int) bool {
		return c[i].Precedence < c[j].Precedence
	})
}

// Check if a bunch of reusable policies defined in K8S are present in an arbirary list of Cloudflare policies
// Ignore all non-reusable, or "imperatively declared elsewhere" policies.
//
//nolint:cyclop
func DoK8SAccessPoliciesMatch(cf *zero_trust.AccessApplicationGetResponse, k8s *v4alpha1.CloudflareAccessApplication) bool {
	if cf == nil && k8s == nil {
		return true
	}

	if cf == nil || k8s == nil {
		return false
	}

	// If the K8s policy list is empty, we consider all policies are present
	if len(*&k8s.Status.ReusablePolicyIDs) == 0 {
		return true
	}

	// cast underlying
	cfPolicies := cf.Policies.([]map[string]any)

	// If the Cloudflare policy list is empty but we have K8s policies,
	// then not all K8s policies are present
	if len(cfPolicies) == 0 {
		return false
	}

	// For each K8s policy, check if its ID exists in the Cloudflare policies
	for naturalPrecedenceIndex, k8sCfPolicyId := range k8s.Status.ReusablePolicyIDs {
		foundMatch := false
		for _, cfPolicy := range cfPolicies {
			// get "id" field
			cfPolicyId, ok := cfPolicy["id"].(string)
			if !ok {
				continue
			}

			// if not the same, skip
			if cfPolicyId != k8sCfPolicyId {
				continue
			}

			//
			// now, we have found a match ! let's check precedence if meaningful
			//

			// get "precedence" field (if any)
			precedence, ok := cfPolicy["precedence"].(int64)

			// if natural precedence is not the same as "precedence", means that it is a no-match
			if ok && int64(naturalPrecedenceIndex+1) != precedence {
				break
			}

			// If IDs match, we also check the precedence
			// (This part can be adapted based on specific needs)
			foundMatch = true
			break
		}

		// If a K8s policy was not found in Cloudflare policies,
		// then not all K8s policies are present
		if !foundMatch {
			return false
		}
	}

	// All K8s policies were found in Cloudflare policies
	return true
}
