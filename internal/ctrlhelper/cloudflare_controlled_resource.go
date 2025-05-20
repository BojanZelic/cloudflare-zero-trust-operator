package ctrlhelper

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Defines a resource that is controlled, with a CloudFlare counterpart
type CloudflareControlledResource interface {
	client.Object

	// Returns the associated CloudFlare's ressource UUID. Might be empty if ressource is not ready or failing.
	GetCloudflareUUID() string

	//
	UnderDeletion() bool

	// Describe the Resource. Most likely the type name. Used for logging or debugging.
	Describe() string

	//
	GetConditions() *[]metav1.Condition

	//
	GetNamespacedName() types.NamespacedName
}
