package ctrlhelper

import (
	"context"
	"strconv"

	v1alpha1meta "github.com/bojanzelic/cloudflare-zero-trust-operator/api/v1alpha1/meta"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logger "sigs.k8s.io/controller-runtime/pkg/log"
)

// type CloudflareCR interface {
// 	*v1alpha1.CloudflareServiceToken | *v1alpha1.CloudflareAccessApplication | *v1alpha1.CloudflareAccessGroup
// }

// func EnsureFinalizer[C CloudflareCR](cr C) {

// }

func EnsureFinalizer(ctx context.Context, r client.Client, c client.Object) error {
	log := logger.FromContext(ctx).WithName("CloudflareAccessGroupController")

	annotations := c.GetAnnotations()
	preventDestroy := false
	if annotationPreventDestroy, ok := annotations[v1alpha1meta.AnnotationPreventDestroy]; ok {
		preventDestroy, _ = strconv.ParseBool(annotationPreventDestroy)
	}

	if preventDestroy && controllerutil.ContainsFinalizer(c, v1alpha1meta.FinalizerDeletion) {
		controllerutil.RemoveFinalizer(c, v1alpha1meta.FinalizerDeletion)
		if err := r.Update(ctx, c); err != nil {
			log.Error(err, "unable to remove finalizer")

			return err
		}
	} else if !controllerutil.ContainsFinalizer(c, v1alpha1meta.FinalizerDeletion) {
		controllerutil.AddFinalizer(c, v1alpha1meta.FinalizerDeletion)
		if err := r.Update(ctx, c); err != nil {
			log.Error(err, "unable to add finalizer")

			return err
		}
	}

	return nil
}
