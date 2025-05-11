package v4alpha1

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/types"
)

type GoogleGroup struct {
	// Google group email
	Email string `json:"email"`
	// An Identity Provider UUID of a "google" type Identity Provider (https://developers.cloudflare.com/api/resources/zero_trust/subresources/identity_providers/methods/get/)
	IdentityProviderID string `json:"identityProviderId"`
}

type OktaGroup struct {
	// Name of the Okta Group
	Name string `json:"name"`
	// An Identity Provider UUID of a "okta" type Identity Provider (https://developers.cloudflare.com/api/resources/zero_trust/subresources/identity_providers/methods/get/)
	IdentityProviderID string `json:"identityProviderId"`
}

type SAMLGroup struct {
	// Name of the SAML group
	Name string `json:"name"`
	// Value of the SAML group
	Value string `json:"value"`
	// An Identity Provider UUID of a "saml" type Identity Provider (https://developers.cloudflare.com/api/resources/zero_trust/subresources/identity_providers/methods/get/)
	IdentityProviderID string `json:"identityProviderId"`
}

type GithubOrganization struct {
	// The name of the organization.
	Name string `json:"name"`
	// The name of the team, if restricting to it.
	// +optional
	Team string `json:"team,omitempty"`
	// An Identity Provider UUID of a "github" type Identity Provider (https://developers.cloudflare.com/api/resources/zero_trust/subresources/identity_providers/methods/get/)
	IdentityProviderID string `json:"identityProviderId"`
}

//
//
//

func parseNamespacedNames(parsableNames []string, contextNamespace string) (nsNames []types.NamespacedName, err error) {
	for _, parsableName := range parsableNames {
		//
		parsed, tErr := _parseNamespacedName(parsableName, contextNamespace)

		// if any failure...
		if tErr != nil {
			err = errors.Wrapf(tErr, "issue while parsing name \"%s\" to namespace", parsableName)
			nsNames = []types.NamespacedName{}
			return
		}

		//
		nsNames = append(nsNames, parsed)
	}

	return
}

// ParseNamespacedName parses a string into a types.NamespacedName.
// Accepts "namespace/name" or just "name" (in which case namespace wil be contextNamespace).
func _parseNamespacedName(s string, contextNamespace string) (types.NamespacedName, error) {
	if s == "" {
		return types.NamespacedName{}, fmt.Errorf("input string is empty")
	}

	parts := strings.SplitN(s, "/", 2)

	switch len(parts) {
	case 1:
		if parts[0] == "" {
			return types.NamespacedName{}, fmt.Errorf("invalid name: cannot be empty")
		}
		return types.NamespacedName{
			Namespace: contextNamespace,
			Name:      parts[0],
		}, nil
	case 2:
		if parts[0] == "" || parts[1] == "" {
			return types.NamespacedName{}, fmt.Errorf("invalid namespaced name: namespace and name must be non-empty")
		}
		return types.NamespacedName{
			Namespace: parts[0],
			Name:      parts[1],
		}, nil
	default:
		return types.NamespacedName{}, fmt.Errorf("unexpected format: %s", s)
	}
}
