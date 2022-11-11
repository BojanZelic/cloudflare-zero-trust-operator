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

// CloudflareAccessGroupSpec defines the desired state of CloudflareAccessGroup.
type CloudflareAccessGroupSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	ZoneID  string                      `json:"zoneId,omitempty"`
	Include []CloudFlareAccessGroupRule `json:"include,omitempty"`
	Require []CloudFlareAccessGroupRule `json:"require,omitempty"`
	Exclude []CloudFlareAccessGroupRule `json:"exclude,omitempty"`
}

type CloudFlareAccessGroupRule struct {
	Emails       []string `json:"emails,omitempty"`
	EmailDomains []string `json:"emailDomains,omitempty"`
	IPRanges     []string `json:"ipRanges,omitempty"`
	// Reference to other access groups
	AccessGroups []string `json:"accessGroups,omitempty"`
	// @todo: add the rest of the fields

	// ValidCertificate []string
	ServiceToken          []string `json:"serviceToken,omitempty"`
	AnyAccessServiceToken *bool    `json:"anyAccessServiceToken,omitempty"`
}

// CloudflareAccessGroupStatus defines the observed state of CloudflareAccessGroup.
type CloudflareAccessGroupStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	AccessGroupID string      `json:"accessGroupId,omitempty"`
	CreatedAt     metav1.Time `json:"createdAt,omitempty"`
	UpdatedAt     metav1.Time `json:"updatedAt,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// CloudflareAccessGroup is the Schema for the cloudflareaccessgroups API.
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
	accessGroup := cloudflare.AccessGroup{
		Name:      c.CloudflareName(),
		ID:        c.Status.AccessGroupID,
		CreatedAt: &c.Status.CreatedAt.Time,
		UpdatedAt: &c.Status.UpdatedAt.Time,
		Include:   make([]interface{}, 0),
		Exclude:   make([]interface{}, 0),
		Require:   make([]interface{}, 0),
	}

	managedCRFields := CloudFlareAccessGroupRuleGroups{
		c.Spec.Include,
		c.Spec.Exclude,
		c.Spec.Require,
	}

	managedCFFields := []*[]interface{}{
		&accessGroup.Include,
		&accessGroup.Exclude,
		&accessGroup.Require,
	}

	managedCRFields.TransformCloudflareRuleFields(managedCFFields)

	return accessGroup
}

type CloudFlareAccessGroupRuleGroups [][]CloudFlareAccessGroupRule

func (c CloudFlareAccessGroupRuleGroups) TransformCloudflareRuleFields(managedCFFields []*[]interface{}) {
	for i, managedField := range c { //nolint:varnamelen
		for _, field := range managedField {
			for _, email := range field.Emails {
				*managedCFFields[i] = append(*managedCFFields[i], cfapi.NewAccessGroupEmail(email))
			}
			for _, domain := range field.EmailDomains {
				*managedCFFields[i] = append(*managedCFFields[i], cfapi.NewAccessGroupEmailDomains(domain))
			}
			for _, ip := range field.IPRanges {
				*managedCFFields[i] = append(*managedCFFields[i], cfapi.NewAccessGroupIP(ip))
			}
			for _, token := range field.ServiceToken {
				*managedCFFields[i] = append(*managedCFFields[i], cfapi.NewAccessGroupServiceToken(token))
			}
			if field.AnyAccessServiceToken != nil && *field.AnyAccessServiceToken {
				*managedCFFields[i] = append(*managedCFFields[i], cfapi.NewAccessGroupAnyValidServiceToken())
			}
			// @todo - make this a reference to another access group instead of an ID
			for _, id := range field.AccessGroups {
				*managedCFFields[i] = append(*managedCFFields[i], cfapi.NewAccessGroupAccessGroup(id))
			}
		}
	}
}

//+kubebuilder:object:root=true

// CloudflareAccessGroupList contains a list of CloudflareAccessGroup.
type CloudflareAccessGroupList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CloudflareAccessGroup `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CloudflareAccessGroup{}, &CloudflareAccessGroupList{})
}
