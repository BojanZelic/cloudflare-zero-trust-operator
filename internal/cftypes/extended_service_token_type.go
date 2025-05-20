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
	var ok bool //nolint:varnamelen

	//
	// Check Annotations
	//
	clientIDKey, ok := secret.Annotations[meta.AnnotationClientIDKey]
	if !ok {
		return ErrMissingAnnotationClientIDKey //nolint:wrapcheck
	}
	clientSecretKey, ok := secret.Annotations[meta.AnnotationClientSecretKey]
	if !ok {
		return ErrMissingAnnotationClientSecretKey //nolint:wrapcheck
	}
	if _, ok = secret.Annotations[meta.AnnotationTokenIDKey]; !ok {
		return ErrMissingAnnotationTokenIDKey //nolint:wrapcheck
	}

	//
	// Check Data
	//
	clientID, ok := secret.Data[clientIDKey]
	if !ok {
		return ErrMissingClientIDKey //nolint:wrapcheck
	}
	clientSecret, ok := secret.Data[clientSecretKey]
	if !ok {
		return ErrMissingClientSecretKey //nolint:wrapcheck
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
