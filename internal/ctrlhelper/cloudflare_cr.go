package ctrlhelper

import "sigs.k8s.io/controller-runtime/pkg/client"

type CloudflareCR interface {
	GetID() string
	GetType() string
	UnderDeletion() bool
	GetStatus() interface{}
	client.Object
}
