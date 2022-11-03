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
	"github.com/bojanzelic/cloudflare-zero-trust-operator/internal/cfapi"
	cloudflare "github.com/cloudflare/cloudflare-go"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// CloudflareAccessGroupSpec defines the desired state of CloudflareAccessGroup
type CloudflareAccessGroupSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	//ZoneID is the cloudflare zone to use
	ZoneID   string                      `json:"zoneId,omitempty"`
	Include  []CloudFlareAccessGroupRule `json:"include,omitempty"`
	Required []CloudFlareAccessGroupRule `json:"required,omitempty"`
	Exclude  []CloudFlareAccessGroupRule `json:"exclude,omitempty"`
}

type CloudFlareAccessGroupRule struct {
	Emails         []string `json:"emails,omitempty"`
	EmailsEndingIn []string `json:"emailsEndingIn,omitempty"`
	IPRanges       []string `json:"ipRanges,omitempty"`
	//AccessGroups   []string
	//Country        []string
	//CommonName     []string
	//ValidCertificate []string
	ServiceToken          []string `json:"serviceToken,omitempty"`
	AnyAccessServiceToken *bool    `json:"anyAccessServiceToken,omitempty"`
}

// CloudflareAccessGroupStatus defines the observed state of CloudflareAccessGroup
type CloudflareAccessGroupStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	AccessGroupID string      `json:"accessGroupId,omitempty"`
	CreatedAt     metav1.Time `json:"createdAt,omitempty"`
	UpdatedAt     metav1.Time `json:"updatedAt,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// CloudflareAccessGroup is the Schema for the cloudflareaccessgroups API
type CloudflareAccessGroup struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CloudflareAccessGroupSpec   `json:"spec,omitempty"`
	Status CloudflareAccessGroupStatus `json:"status,omitempty"`
}

func (c CloudflareAccessGroup) CloudflareName() string {
	return c.ObjectMeta.Name + " [K8s]"
}

func (c *CloudflareAccessGroup) ToCloudflare() cloudflare.AccessGroup {
	ag := cloudflare.AccessGroup{
		Name:      c.CloudflareName(),
		ID:        c.Status.AccessGroupID,
		CreatedAt: &c.Status.CreatedAt.Time,
		UpdatedAt: &c.Status.UpdatedAt.Time,
		Include:   make([]interface{}, 0),
		Exclude:   make([]interface{}, 0),
		Require:   make([]interface{}, 0),
	}

	for _, include := range c.Spec.Include {
		for _, email := range include.Emails {
			ag.Include = append(ag.Include, cfapi.NewAccessGroupEmail(email))
		}
	}

	return ag
}

//+kubebuilder:object:root=true

// CloudflareAccessGroupList contains a list of CloudflareAccessGroup
type CloudflareAccessGroupList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CloudflareAccessGroup `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CloudflareAccessGroup{}, &CloudflareAccessGroupList{})
}
