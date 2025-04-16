package services

import (
	"context"

	"github.com/bojanzelic/cloudflare-zero-trust-operator/api/v1alpha1"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type AccessPolicyService struct {
	Client client.Client
	Log    logr.Logger
}

type AccessPolicyList interface {
	GetInclude() []v1alpha1.CloudFlareAccessGroupRule
	GetExclude() []v1alpha1.CloudFlareAccessGroupRule
	GetRequire() []v1alpha1.CloudFlareAccessGroupRule
}

func ToLegacyAccessPolicyList(abs v1alpha1.CloudflareLegacyAccessPolicyList) []AccessPolicyList {
	result := make([]AccessPolicyList, 0, len(abs))
	for _, policy := range abs {
		result = append(result, policy)
	}

	return result
}

// nolint: gocognit
func (s *AccessPolicyService) PopulateLegacyAccessPolicyReferences(ctx context.Context, policyList []AccessPolicyList) error {
	for _, policy := range policyList {
		include := policy.GetInclude()
		exclude := policy.GetExclude()
		require := policy.GetRequire()

		managedCRFields := []*[]v1alpha1.CloudFlareAccessGroupRule{
			&include,
			&exclude,
			&require,
		}

		for _, fields := range managedCRFields {
			for j, field := range *fields {
				for k, token := range field.AccessGroups {
					if token.ValueFrom != nil {
						accessGroup := &v1alpha1.CloudflareAccessGroup{}
						if err := s.Client.Get(ctx, token.ValueFrom.ToNamespacedName(), accessGroup); err != nil {
							return errors.Wrapf(err, "unable to reference CloudflareAccessGroup %s - %s", token.ValueFrom.Name, token.ValueFrom.Namespace)
						}

						(*fields)[j].AccessGroups[k].Value = accessGroup.Status.AccessGroupID
					}
				}

				for k, token := range field.ServiceToken {
					if token.ValueFrom != nil {
						serviceToken := &v1alpha1.CloudflareServiceToken{}
						if err := s.Client.Get(ctx, token.ValueFrom.ToNamespacedName(), serviceToken); err != nil {
							return errors.Wrapf(err, "unable to reference CloudflareServiceToken %s - %s", token.ValueFrom.Name, token.ValueFrom.Namespace)
						}

						(*fields)[j].ServiceToken[k].Value = serviceToken.Status.ServiceTokenID
					}
				}
			}
		}
	}

	return nil
}
