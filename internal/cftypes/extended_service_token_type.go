package cftypes

import (
	"errors"

	"github.com/cloudflare/cloudflare-go"
	corev1 "k8s.io/api/core/v1"
)

var (
	ErrMissingClientIDKey     = errors.New("missing clientIDKey field in secret data")
	ErrMissingClientSecretKey = errors.New("missing clientSecretKey field in secret data")
)

type ExtendedServiceToken struct {
	cloudflare.AccessServiceToken
	ClientSecret string
	K8sSecretRef struct {
		ClientIDKey     string
		ClientSecretKey string
		SecretName      string
	}
}

func (s *ExtendedServiceToken) SetSecretValues(clientIDKey, clientSecretKey string, secret corev1.Secret) error {
	if _, ok := secret.Data[clientIDKey]; !ok {
		return ErrMissingClientIDKey
	}

	if _, ok := secret.Data[clientSecretKey]; !ok {
		return ErrMissingClientSecretKey
	}

	s.ClientID = string(secret.Data[clientIDKey])
	s.ClientSecret = string(secret.Data[clientSecretKey])

	s.SetSecretReference(clientIDKey, clientSecretKey, secret)

	return nil
}

// depricated.
func (s *ExtendedServiceToken) SetSecretReference(clientIDKey, clientSecretKey string, secret corev1.Secret) {
	s.K8sSecretRef.ClientIDKey = clientIDKey
	s.K8sSecretRef.ClientSecretKey = clientSecretKey
	s.K8sSecretRef.SecretName = secret.Name
}
