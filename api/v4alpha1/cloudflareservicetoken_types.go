/*
Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v4alpha1

import (
	"github.com/bojanzelic/cloudflare-zero-trust-operator/internal/cftypes"
	"github.com/cloudflare/cloudflare-go/v4/zero_trust"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// CloudflareServiceTokenSpec defines the desired state of CloudflareServiceToken.
type CloudflareServiceTokenSpec struct {
	// Name of the Cloudflare Access Group
	Name string `json:"name"`

	// Time before the token should be automatically renewed. Defaults to "0"
	// Automatically renewing a service token will change the service token value upon renewal.
	// Tokens will get automatically renewed if the token is expired
	//
	// +optional
	// +kubebuilder:default="0"
	MinTimeBeforeRenewal string `json:"minTimeBeforeRenewal,omitzero"`

	// Recreate the token if the secret with the service token value is missing or doesn't exist
	//
	// +kubebuilder:default=true
	RecreateMissing bool `json:"recreateMissing,omitzero"`

	// Template to apply for the generated secret
	//
	// +optional
	// +kubebuilder:default={"metadata": {}}
	Template SecretTemplateSpec `json:"template"`
}

type SecretTemplateSpec struct {
	// Standard object's metadata.
	// More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata
	//
	// +optional
	// +nullable
	// +kubebuilder:validation:XPreserveUnknownFields
	metav1.ObjectMeta `json:"metadata" protobuf:"bytes,1,opt,name=metadata"`

	// Key that should store the secret data. Defaults to cloudflareServiceToken
	//
	// Warning: changing this value will recreate the secret
	//
	// +optional
	// +kubebuilder:default=cloudflareSecretKey
	ClientSecretKey string `json:"clientSecretKey,omitzero"`

	// Key that should store the secret data. Defaults to cloudflareServiceToken.
	//
	// Warning: changing this value will recreate the secret
	//
	// +optional
	// +kubebuilder:default=cloudflareClientId
	ClientIDKey string `json:"clientIdKey,omitzero"`
}

// CloudflareServiceTokenStatus defines the observed state of CloudflareServiceToken.
type CloudflareServiceTokenStatus struct {
	// ID of the servicetoken in Cloudflare
	ServiceTokenID string `json:"serviceTokenId,omitzero"`

	// Creation timestamp of the resource in Cloudflare
	CreatedAt metav1.Time `json:"createdAt,omitzero"`

	// Updated timestamp of the resource in Cloudflare
	UpdatedAt metav1.Time `json:"updatedAt,omitzero"`

	// Updated timestamp of the resource in Cloudflare
	ExpiresAt metav1.Time `json:"expiresAt,omitzero"`

	// SecretRef is the reference to the secret
	//
	// +optional
	SecretRef SecretRef `json:"secretRef,omitzero"`

	// Conditions store the status conditions of the CloudflareAccessApplication
	//
	// +operator-sdk:csv:customresourcedefinitions:type=status
	Conditions []metav1.Condition `json:"conditions,omitempty" patchMergeKey:"type" patchStrategy:"merge" protobuf:"bytes,1,rep,name=conditions"`
}

type SecretRef struct {
	// reference to the secret
	corev1.LocalObjectReference `json:"reference,omitzero"`
	// Key that stores the secret data.
	ClientSecretKey string `json:"clientSecretKey,omitzero"`
	// Key that stores the secret data.
	ClientIDKey string `json:"clientIdKey,omitzero"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// CloudflareServiceToken is the Schema for the cloudflareservicetokens API.
type CloudflareServiceToken struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitzero"`

	Spec   CloudflareServiceTokenSpec   `json:"spec,omitzero"`
	Status CloudflareServiceTokenStatus `json:"status,omitzero"`
}

func (c *CloudflareServiceToken) GetType() string {
	return "CloudflareServiceToken"
}

func (c *CloudflareServiceToken) GetID() string {
	return c.Status.ServiceTokenID
}

func (c *CloudflareServiceToken) UnderDeletion() bool {
	return !c.DeletionTimestamp.IsZero()
}

func (c *CloudflareServiceToken) ToExtendedToken() cftypes.ExtendedServiceToken {
	return cftypes.ExtendedServiceToken{
		ServiceToken: zero_trust.ServiceToken{
			CreatedAt: c.Status.CreatedAt.Time,
			UpdatedAt: c.Status.UpdatedAt.Time,
			ExpiresAt: c.Status.ExpiresAt.Time,
			ID:        c.Status.ServiceTokenID,
			Name:      c.Spec.Name,
		},
	}
}

// +kubebuilder:object:root=true

// CloudflareServiceTokenList contains a list of CloudflareServiceToken.
type CloudflareServiceTokenList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitzero"`
	Items           []CloudflareServiceToken `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CloudflareServiceToken{}, &CloudflareServiceTokenList{})
}
