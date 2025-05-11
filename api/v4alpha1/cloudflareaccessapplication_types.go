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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// CloudflareAccessApplicationSpec defines the desired state of CloudflareAccessApplication.
type CloudflareAccessApplicationSpec struct {
	// The application type. If omitted, resolves to "self_hosted". Only a bunch of official types are supported.
	//
	// https://developers.cloudflare.com/api/resources/zero_trust/subresources/access/subresources/applications/models/application_type/
	//
	// +optional
	// +kubebuilder:default=self_hosted
	// +kubebuilder:validation:Enum=self_hosted;warp;app_launcher
	Type string `json:"type,omitempty"`

	// Name of the Cloudflare Access Application.
	//
	// Meaningless for "warp" and "app_launcher" app types. Required for "self_hosted".
	//
	// +optional
	Name string `json:"name,omitempty"`

	// The domain and path that Access will secure.
	//
	// Meaningless for "warp" and "app_launcher" app types. Required for "self_hosted".
	//
	// ex: "test.example.com/admin"
	//
	// +optional
	Domain string `json:"domain,omitempty"`

	// Specify if the application will be visible in the App Launcher.
	//
	// Meaningless for "warp" and "app_launcher" app types.
	//
	// +optional
	// +kubebuilder:default=true
	AppLauncherVisible *bool `json:"appLauncherVisible,omitempty"`

	// The identity providers your users can select when connecting to this application. Defaults to all IdPs configured in your account.
	//
	// ex: ["699d98642c564d2e855e9661899b7252"]
	//
	// +optional
	// +kubebuilder:default={}
	AllowedIdps []string `json:"allowedIdps,omitempty"`

	// When set to true, users skip the identity provider selection step during login.
	// You must specify only one identity provider in allowed_idps.
	//
	// +optional
	// +kubebuilder:default=false
	AutoRedirectToIdentity *bool `json:"autoRedirectToIdentity,omitempty"`

	// PolicyRefs is an ordered slice of names or {namespace/name} referencing [CloudflareAccessReusablePolicy].
	// Referenced policies would be applied to this access application.
	// Order determines precedence
	//
	// +optional
	PolicyRefs []string `json:"policyRefs,omitempty"`

	// SessionDuration is the length of the session duration.
	//
	// +optional
	// +kubebuilder:default="24h"
	SessionDuration string `json:"sessionDuration,omitempty"`

	// Enables the binding cookie, which increases security against compromised authorization tokens and CSRF attacks.
	//
	// Meaningless for "warp" and "app_launcher" app types.
	//
	// +optional
	// +kubebuilder:default=false
	EnableBindingCookie *bool `json:"enableBindingCookie,omitempty"`

	// Enables the HttpOnly cookie attribute, which increases security against XSS attacks.
	//
	// Meaningless for "warp" and "app_launcher" app types.
	//
	// +optional
	// +kubebuilder:default=true
	HTTPOnlyCookieAttribute *bool `json:"httpOnlyCookieAttribute,omitempty"`

	// The image URL for the logo shown in the App Launcher dashboard
	//
	// +optional
	LogoURL string `json:"logoUrl,omitempty"`
}

func (spec *CloudflareAccessApplicationSpec) GetNamespacedPolicyRefs(contextNamespace string) ([]types.NamespacedName, error) {
	return parseNamespacedNames(spec.PolicyRefs, contextNamespace)
}

// CloudflareAccessApplicationStatus defines the observed state of CloudflareAccessApplication.
type CloudflareAccessApplicationStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	AccessApplicationID string `json:"accessApplicationId,omitempty"`

	// ordered CloudFlare's policies IDs, resolved by controller from "Spec.PolicyRefs"
	ReusablePolicyIDs []string    `json:"reusablePolicyIds,omitempty"`
	CreatedAt         metav1.Time `json:"createdAt"`
	UpdatedAt         metav1.Time `json:"updatedAt"`

	// Conditions store the status conditions of the CloudflareAccessApplication
	//
	// +operator-sdk:csv:customresourcedefinitions:type=status
	Conditions []metav1.Condition `json:"conditions,omitempty" patchMergeKey:"type" patchStrategy:"merge" protobuf:"bytes,1,rep,name=conditions"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// CloudflareAccessApplication is the Schema for the cloudflareaccessapplications API.
type CloudflareAccessApplication struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`

	Spec   CloudflareAccessApplicationSpec   `json:"spec"`
	Status CloudflareAccessApplicationStatus `json:"status"`
}

func (c *CloudflareAccessApplication) GetType() string {
	return "CloudflareAccessApplication"
}

func (c *CloudflareAccessApplication) GetID() string {
	return c.Status.AccessApplicationID
}

func (c *CloudflareAccessApplication) UnderDeletion() bool {
	return !c.DeletionTimestamp.IsZero()
}

// +kubebuilder:object:root=true

// CloudflareAccessApplicationList contains a list of CloudflareAccessApplication.
type CloudflareAccessApplicationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []CloudflareAccessApplication `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CloudflareAccessApplication{}, &CloudflareAccessApplicationList{})
}
