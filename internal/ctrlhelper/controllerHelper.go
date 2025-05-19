package ctrlhelper

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/Southclaws/fault"
	"github.com/Southclaws/fault/fmsg"
	"github.com/bojanzelic/cloudflare-zero-trust-operator/api/v4alpha1"
	"github.com/bojanzelic/cloudflare-zero-trust-operator/internal/cfapi"
	"github.com/bojanzelic/cloudflare-zero-trust-operator/internal/meta"
	cloudflare "github.com/cloudflare/cloudflare-go/v4"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

type ControllerHelper struct {
	R client.Client

	// Determines requeue delays on "normal", errorless requeue attempt, mostly for awaiting other ressources to be ready.
	//
	// You might want to get that lower on tests
	NormalRequeueDelay time.Duration
}

func (h *ControllerHelper) ensureFinalizer(
	ctx context.Context,
	c CloudflareControlledResource, //nolint:varnamelen
) error {
	annotations := c.GetAnnotations()
	preventDestroy := false
	if annotationPreventDestroy, ok := annotations[meta.AnnotationPreventDestroy]; ok {
		preventDestroy, _ = strconv.ParseBool(annotationPreventDestroy)
	}

	if preventDestroy && controllerutil.ContainsFinalizer(c, meta.FinalizerDeletion) {
		controllerutil.RemoveFinalizer(c, meta.FinalizerDeletion)
		if err := h.R.Update(ctx, c); err != nil {
			return fault.Wrap(err, fmsg.With("unable to remove finalizer"))
		}
	} else if !preventDestroy && !controllerutil.ContainsFinalizer(c, meta.FinalizerDeletion) {
		controllerutil.AddFinalizer(c, meta.FinalizerDeletion)
		if err := h.R.Update(ctx, c); err != nil {
			return fault.Wrap(err, fmsg.With("unable to add finalizer"))
		}
	}

	return nil
}

//nolint:cyclop
func (h *ControllerHelper) ReconcileDeletion(ctx context.Context, api *cfapi.API, k8sCR CloudflareControlledResource) (bool, error) {
	log := ctrl.LoggerFrom(ctx).WithName("finalizerHelper::ReconcileDeletion").WithValues(
		"type", k8sCR.GetObjectKind().GroupVersionKind().Kind,
		"name", k8sCR.GetName(),
		"namespace", k8sCR.GetNamespace(),
	)

	k8sCR.GetObjectKind()

	// examine DeletionTimestamp to determine if object is under deletion
	if !k8sCR.UnderDeletion() {
		if err := h.ensureFinalizer(ctx, k8sCR); err != nil {
			return false, fault.Wrap(err, fmsg.With("unable to reconcile finalizer"))
		}

		return true, nil
	}

	// The object is being deleted
	//nolint:nestif
	if controllerutil.ContainsFinalizer(k8sCR, meta.FinalizerDeletion) {
		// our finalizer is present, so lets handle any external dependency
		if k8sCR.GetCloudflareUUID() != "" {
			log.Info("will remove resource in Cloudflare")
			var err error

			switch k8sCR.(type) {
			case *v4alpha1.CloudflareAccessApplication:
				err = api.DeleteAccessApplication(ctx, k8sCR.GetCloudflareUUID())
			case *v4alpha1.CloudflareAccessGroup:
				err = api.DeleteAccessGroup(ctx, k8sCR.GetCloudflareUUID())
			case *v4alpha1.CloudflareServiceToken:
				err = api.DeleteAccessServiceToken(ctx, k8sCR.GetCloudflareUUID())
			case *v4alpha1.CloudflareAccessReusablePolicy:
				err = api.DeleteAccessReusablePolicy(ctx, k8sCR.GetCloudflareUUID())
			default:
				return false, fault.Newf("unknown type %T", k8sCR)
			}

			if err != nil {
				var cfErr *cloudflare.Error
				if errors.As(err, &cfErr) && cfErr.StatusCode == 404 {
					log.Info("unable to remove resource from cloudflare - appears to be already deleted")
				} else {
					return false, fault.Wrap(err, fmsg.With("unable to delete"))
				}
			} else {
				log.Info("resource removed in Cloudflare")
			}
		}

		// remove our finalizer from the list and update it.
		controllerutil.RemoveFinalizer(k8sCR, meta.FinalizerDeletion)
		if err := h.R.Update(ctx, k8sCR); err != nil {
			return false, fault.Wrap(err, fmsg.With("unable to remove finalizer"))
		}
	}

	// Stop reconciliation as the item is being deleted
	log.Info("destroyed successfully")

	return false, nil
}
