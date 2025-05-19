package ctrlhelper

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Defines a resource that is controlled, with a CloudFlare counterpart
type CloudflareControlledResource interface {
	client.Object

	// Returns the associated CloudFlare's ressource UUID. Might be empty if ressource is not ready or failing.
	GetCloudflareUUID() string

	//
	UnderDeletion() bool

	//
	GetConditions() *[]metav1.Condition
}
