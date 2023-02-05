package ctrlhelper

import (
	"context"
	"strconv"

	"github.com/bojanzelic/cloudflare-zero-trust-operator/api/v1alpha1"
	v1alpha1meta "github.com/bojanzelic/cloudflare-zero-trust-operator/api/v1alpha1/meta"
	"github.com/bojanzelic/cloudflare-zero-trust-operator/internal/cfapi"
	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logger "sigs.k8s.io/controller-runtime/pkg/log"
)

type ControllerHelper struct {
	R client.Client
}

func (h *ControllerHelper) EnsureFinalizer(ctx context.Context, c CloudflareCR) error {
	log := logger.FromContext(ctx).WithName("finalizerHelper::CloudflareAccessGroupController")

	annotations := c.GetAnnotations()
	preventDestroy := false
	if annotationPreventDestroy, ok := annotations[v1alpha1meta.AnnotationPreventDestroy]; ok {
		preventDestroy, _ = strconv.ParseBool(annotationPreventDestroy)
	}

	if preventDestroy && controllerutil.ContainsFinalizer(c, v1alpha1meta.FinalizerDeletion) {
		controllerutil.RemoveFinalizer(c, v1alpha1meta.FinalizerDeletion)
		if err := h.R.Update(ctx, c); err != nil {
			log.Error(err, "unable to remove finalizer")

			return err
		}
	} else if !controllerutil.ContainsFinalizer(c, v1alpha1meta.FinalizerDeletion) {
		controllerutil.AddFinalizer(c, v1alpha1meta.FinalizerDeletion)
		if err := h.R.Update(ctx, c); err != nil {
			log.Error(err, "unable to add finalizer")

			return err
		}
	}

	return nil
}

func (r *ControllerHelper) ReconcileDeletion(ctx context.Context, api *cfapi.API, k8sCR CloudflareCR) (bool, error) {
	log := logger.FromContext(ctx).WithName("finalizerHelper::ReconcileDeletion").WithValues(map[string]string{
		"type":      k8sCR.GetType(),
		"name":      k8sCR.GetName(),
		"namespace": k8sCR.GetNamespace(),
	})

	// examine DeletionTimestamp to determine if object is under deletion
	if !k8sCR.UnderDeletion() {
		if err := r.EnsureFinalizer(ctx, k8sCR); err != nil {
			return false, errors.Wrap(err, "unable to reconcile finalizer")
		}
	} else {
		// The object is being deleted
		if controllerutil.ContainsFinalizer(k8sCR, v1alpha1meta.FinalizerDeletion) {
			// our finalizer is present, so lets handle any external dependency
			if k8sCR.GetID() != "" {
				var err error

				switch k8sCR.(type) {
				case *v1alpha1.CloudflareAccessApplication:
					err = api.DeleteAccessApplication(ctx, k8sCR.GetID())
				case *v1alpha1.CloudflareAccessGroup:
					err = api.DeleteAccessGroup(ctx, k8sCR.GetID())
				case *v1alpha1.CloudflareServiceToken:
					err = api.DeleteAccessServiceToken(ctx, k8sCR.GetID())
				default:
					return false, errors.Errorf("unknown type %T", k8sCR)
				}

				if err != nil {
					log.Error(err, "unable to delete")

					return false, errors.Wrap(err, "unable to delete")
				}
			}

			// remove our finalizer from the list and update it.
			controllerutil.RemoveFinalizer(k8sCR, v1alpha1meta.FinalizerDeletion)
			if err := r.R.Update(ctx, k8sCR); err != nil {
				log.Error(err, "unable to remove finalizer")

				return false, err
			}
		}

		// Stop reconciliation as the item is being deleted
		log.Info("destroyed successfully")

		return false, nil
	}

	return true, nil
}
