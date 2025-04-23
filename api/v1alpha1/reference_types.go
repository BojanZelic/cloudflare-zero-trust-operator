package v1alpha1

import "k8s.io/apimachinery/pkg/types"

type AccessGroup struct {
	// Optional: no more than one of the following may be specified.
	// ID of the CloudflareAccessGroup
	// +optional
	Value string `json:"value,omitempty" protobuf:"bytes,1,opt,name=value"`
	// Source for the CloudflareAccessGroup's variable. Cannot be used if value is not empty.
	// +optional
	ValueFrom *AccessGroupReference `json:"valueFrom,omitempty" protobuf:"bytes,2,opt,name=valueFrom"`
}

type ServiceToken struct {
	// Optional: no more than one of the following may be specified.
	// ID of the CloudflareServiceToken
	// +optional
	Value string `json:"value,omitempty" protobuf:"bytes,1,opt,name=value"`
	// Source for the CloudflareServiceToken's variable. Cannot be used if value is not empty.
	// +optional
	ValueFrom *ServiceTokenReference `json:"valueFrom,omitempty" protobuf:"bytes,2,opt,name=valueFrom"`
}

type GoogleGroup struct {
	// Google group email
	Email string `json:"email"`
	// Identity Provider Id
	IdentityProviderID string `json:"identityProviderId"`
}

type OktaGroup struct {
	// Name of the Okta Group
	Name string `json:"name"`
	// Identity Provider Id
	IdentityProviderID string `json:"identityProviderId"`
}

type OIDCClaim struct {
	// Name of the OIDC claim
	Name string `json:"name"`
	// Value of the OIDC claim
	Value string `json:"value"`
	// Identity Provider Id
	IdentityProviderID string `json:"identityProviderId"`
}

type GithubOrganization struct {
	// The name of the organization.
	Name string `json:"name"`
	// The ID of your Github identity provider.
	IdentityProviderID string `json:"identityProviderId"`
	// The name of the team
	// +optional
	Team string `json:"team,omitempty"`
}

type AccessGroupReference struct {
	// `namespace` is the namespace of the AccessGroup.
	// Required
	Namespace string `json:"namespace" protobuf:"bytes,1,opt,name=namespace"`
	// `name` is the name of the AccessGroup .
	// Required
	Name string `json:"name" protobuf:"bytes,2,opt,name=name"`
}

func (g *AccessGroupReference) ToNamespacedName() types.NamespacedName {
	return types.NamespacedName{Namespace: g.Namespace, Name: g.Name}
}

type ServiceTokenReference struct {
	// `namespace` is the namespace of the AccessGroup.
	// Required
	Namespace string `json:"namespace" protobuf:"bytes,1,opt,name=namespace"`
	// `name` is the name of the AccessGroup .
	// Required
	Name string `json:"name" protobuf:"bytes,2,opt,name=name"`
}

func (g *ServiceTokenReference) ToNamespacedName() types.NamespacedName {
	return types.NamespacedName{Namespace: g.Namespace, Name: g.Name}
}
