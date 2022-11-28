package cftypes

import (
	corev1 "k8s.io/api/core/v1"
)

// nolint: gosec
const (
	AnnotationClientIDKey     = "cloudflare.zelic.io/client-id-key"
	AnnotationClientSecretKey = "cloudflare.zelic.io/client-secret-key"
	AnnotationTokenIDKey      = "cloudflare.zelic.io/token-id-key"
	LabelOwnedBy              = "cloudflare.zelic.io/owned-by"
)

type ServiceTokenSecret struct {
	corev1.Secret
}
