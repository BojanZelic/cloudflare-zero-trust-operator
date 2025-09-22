package cfcompare

import (
	"context"
	"slices"

	"github.com/Southclaws/fault"
	"github.com/bojanzelic/cloudflare-zero-trust-operator/api/v4alpha1"
	"github.com/bojanzelic/cloudflare-zero-trust-operator/internal/cfapi"
	"github.com/cloudflare/cloudflare-go/v4/zero_trust"
	"github.com/go-logr/logr"
)

//nolint:cyclop,varnamelen
func DoCFPoliciesEquateToK8Ss(
	ctx context.Context,
	log *logr.Logger,
	cf *zero_trust.AccessApplicationGetResponse,
	k8s *v4alpha1.CloudflareAccessApplication,
) bool {
	// if both nil, they are equivalent
	if cf == nil && k8s == nil {
		return true
	}

	// then if any of those is nil and the other is not, they are obviously different
	if cf == nil || k8s == nil {
		return false
	}

	//
	cfPolicyUUIDs, err := cfapi.GetOrderedPolicyUUIDs(cf)
	if err != nil {
		//
		log.Error(fault.New("Issue while enumerating policies, on access application comparaison"), "unable to produce ordered policy UUIDs")

		// just say they match so that we do not trigger infinite update loops
		return true
	}

	areEq := slices.Equal(cfPolicyUUIDs, k8s.Status.ReusablePolicyIDs)
	return areEq
}
