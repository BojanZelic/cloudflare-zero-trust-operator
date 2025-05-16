package cftypes

import (
	"github.com/Southclaws/fault"
	"github.com/bojanzelic/cloudflare-zero-trust-operator/internal/meta"
	"github.com/cloudflare/cloudflare-go/v4/zero_trust"
	corev1 "k8s.io/api/core/v1"
)

var (
	ErrMissingClientIDKey               = fault.New("missing clientIDKey field in secret data")
	ErrMissingClientSecretKey           = fault.New("missing clientSecretKey field in secret data")
	ErrMissingTokenIDKey                = fault.New("missing TokenID field in secret data")
	ErrMissingAnnotationClientIDKey     = fault.New("missing clientIDKey annotation in secret")
	ErrMissingAnnotationClientSecretKey = fault.New("missing clientSecretKey annotation in secret")
	ErrMissingAnnotationTokenIDKey      = fault.New("missing TokenID annotation in secret")
)

type ExtendedServiceToken struct {
	zero_trust.ServiceToken
	ClientSecret string
	K8sSecretRef struct {
		ClientIDKey     string
		ClientSecretKey string
		SecretName      string
	}
}

// Updates ExtendedServiceToken with values
func (st *ExtendedServiceToken) SetSecretValues(secret corev1.Secret) error {
	//
	// Check Annotations
	//
	clientIDKey, ok := secret.Annotations[meta.AnnotationClientIDKey] //nolint:varnamelen
	if !ok {
		return ErrMissingAnnotationClientIDKey
	}
	clientSecretKey, ok := secret.Annotations[meta.AnnotationClientSecretKey] //nolint:varnamelen
	if !ok {
		return ErrMissingAnnotationClientSecretKey
	}
	if _, ok := secret.Annotations[meta.AnnotationTokenIDKey]; !ok { //nolint:varnamelen
		return ErrMissingAnnotationTokenIDKey
	}

	//
	// Check Data
	//
	clientID, ok := secret.Data[clientIDKey] //nolint:varnamelen
	if !ok {
		return ErrMissingClientIDKey
	}
	clientSecret, ok := secret.Data[clientSecretKey] //nolint:varnamelen
	if !ok {
		return ErrMissingClientSecretKey
	}

	//
	// Bind
	//

	st.ClientID = string(clientID)
	st.ClientSecret = string(clientSecret)
	st.SetSecretReference(clientIDKey, clientSecretKey, secret)

	return nil
}

// deprecated.
func (st *ExtendedServiceToken) SetSecretReference(clientIDKey, clientSecretKey string, secret corev1.Secret) {
	st.K8sSecretRef.ClientIDKey = clientIDKey
	st.K8sSecretRef.ClientSecretKey = clientSecretKey
	st.K8sSecretRef.SecretName = secret.Name
}
