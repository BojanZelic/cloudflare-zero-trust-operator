package meta

//nolint:gosec
const (
	AnnotationClientIDKey     = "cloudflare.zelic.io/client-id-key"
	AnnotationClientSecretKey = "cloudflare.zelic.io/client-secret-key"
	AnnotationTokenIDKey      = "cloudflare.zelic.io/token-id-key"
	LabelOwnedBy              = "cloudflare.zelic.io/owned-by"
	FinalizerDeletion         = "cloudflare.zelic.io/finalizer"
	AnnotationPreventDestroy  = "cloudflare.zelic.io/prevent-destroy"
)
