/*
Copyright 2022.

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

package v1alpha1

import (
	cloudflare "github.com/cloudflare/cloudflare-go"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// CloudflareAccessApplicationSpec defines the desired state of CloudflareAccessApplication
type CloudflareAccessApplicationSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	Domain                  string                           `json:"domain"`
	Type                    cloudflare.AccessApplicationType `json:"type,omitempty"`
	AppLauncherVisible      *bool                            `json:"app_launcher_visible,omitempty"`
	AllowedIdps             []string                         `json:"allowed_idps,omitempty"`
	AutoRedirectToIdentity  *bool                            `json:"auto_redirect_to_identity,omitempty"`
	Policies                []*CloudflareAccessPolicy        `json:"policies,omitempty"`
	SessionDuration         string                           `json:"session_duration,omitempty"`
	EnableBindingCookie     *bool                            `json:"enable_binding_cookie,omitempty"`
	HttpOnlyCookieAttribute *bool                            `json:"http_only_cookie_attribute,omitempty"`
}

type CloudflareAccessPolicy struct {
	Name     string                      `json:"name"`
	Decision string                      `json:"decision"`
	Include  []CloudFlareAccessGroupRule `json:"include,omitempty"`
	Require  []CloudFlareAccessGroupRule `json:"require,omitempty"`
	Exclude  []CloudFlareAccessGroupRule `json:"exclude,omitempty"`

	// PurposeJustificationRequired *bool                 `json:"purpose_justification_required,omitempty"`
	// PurposeJustificationPrompt   *string               `json:"purpose_justification_prompt,omitempty"`
	// ApprovalRequired             *bool                 `json:"approval_required,omitempty"`
	// ApprovalGroups               []cloudflare.AccessApprovalGroup `json:"approval_groups"`
}

// CloudflareAccessApplicationStatus defines the observed state of CloudflareAccessApplication
type CloudflareAccessApplicationStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	AccessApplicationID string      `json:"accessApplicationId,omitempty"`
	CreatedAt           metav1.Time `json:"createdAt,omitempty"`
	UpdatedAt           metav1.Time `json:"updatedAt,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// CloudflareAccessApplication is the Schema for the cloudflareaccessapplications API
type CloudflareAccessApplication struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CloudflareAccessApplicationSpec   `json:"spec,omitempty"`
	Status CloudflareAccessApplicationStatus `json:"status,omitempty"`
}

func (c *CloudflareAccessApplication) CloudflareName() string {
	return c.ObjectMeta.Name + " [K8s]"
}

//+kubebuilder:object:root=true

// CloudflareAccessApplicationList contains a list of CloudflareAccessApplication
type CloudflareAccessApplicationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CloudflareAccessApplication `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CloudflareAccessApplication{}, &CloudflareAccessApplicationList{})
}
