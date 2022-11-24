package cftypes

import (
	"errors"

	"github.com/cloudflare/cloudflare-go"
	corev1 "k8s.io/api/core/v1"
)

type ExtendedServiceToken struct {
	cloudflare.AccessServiceToken
	ClientSecret string
}

func (s *ExtendedServiceToken) SetSecretValues(clientIdKey, clientSecretKey string, secret corev1.Secret) error {
	if _, ok := secret.StringData[clientIdKey]; !ok {
		return errors.New("missing clientIdKey field in secret data")
	}

	if _, ok := secret.StringData[clientSecretKey]; !ok {
		return errors.New("missing clientSecretKey field in secret data")
	}

	s.ClientID = secret.StringData[clientIdKey]
	s.ClientSecret = secret.StringData[clientSecretKey]
	return nil
}
