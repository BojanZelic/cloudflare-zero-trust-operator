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
//
//nolint:cyclop
func AreK8SAccessPoliciesPresent(cf *[]zero_trust.AccessApplicationPolicyListResponse, k8s *[]v4alpha1.CloudflareAccessReusablePolicy) bool {
	if cf == nil && k8s == nil {
		return true
	}

	if cf == nil || k8s == nil {
		return false
	}

	// If the K8s policy list is empty, we consider all policies are present
	if len(*k8s) == 0 {
		return true
	}

	// If the Cloudflare policy list is empty but we have K8s policies,
	// then not all K8s policies are present
	if len(*cf) == 0 {
		return false
	}

	// For each K8s policy, check if its ID exists in the Cloudflare policies
	for _, k8sPolicy := range *k8s {
		found := false
		for _, cfPolicy := range *cf {
			// We only compare IDs and precedence
			if k8sPolicy.Status.AccessReusablePolicyID == cfPolicy.ID {
				// If IDs match, we also check the precedence
				// (This part can be adapted based on specific needs)
				found = true
				break
			}
		}

		// If a K8s policy was not found in Cloudflare policies,
		// then not all K8s policies are present
		if !found {
			return false
		}
	}

	// All K8s policies were found in Cloudflare policies
	return true
}
