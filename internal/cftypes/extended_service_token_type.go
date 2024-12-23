package cftypes

import (
	"errors"

	"github.com/cloudflare/cloudflare-go"
	corev1 "k8s.io/api/core/v1"
)

var (
	ErrMissingClientIDKey               = errors.New("missing clientIDKey field in secret data")
	ErrMissingClientSecretKey           = errors.New("missing clientSecretKey field in secret data")
	ErrMissingTokenIDKey                = errors.New("missing TokenID field in secret data")
	ErrMissingAnnotationClientIDKey     = errors.New("missing clientIDKey annotation in secret")
	ErrMissingAnnotationClientSecretKey = errors.New("missing clientSecretKey annotation in secret")
	ErrMissingAnnotationTokenIDKey      = errors.New("missing TokenID annotation in secret")
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

func (s *ExtendedServiceToken) SetSecretValues(secret corev1.Secret) error {
	if _, ok := secret.Annotations["cloudflare.kadaan.info/client-id-key"]; !ok {
		return ErrMissingAnnotationClientIDKey
	}

	if _, ok := secret.Annotations["cloudflare.kadaan.info/client-secret-key"]; !ok {
		return ErrMissingAnnotationClientSecretKey
	}

	if _, ok := secret.Annotations["cloudflare.kadaan.info/token-id-key"]; !ok {
		return ErrMissingAnnotationTokenIDKey
	}

	clientIDKey := secret.Annotations["cloudflare.kadaan.info/client-id-key"]
	clientSecretKey := secret.Annotations["cloudflare.kadaan.info/client-secret-key"]

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
