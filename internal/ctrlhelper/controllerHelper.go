package ctrlhelper

import (
	"context"
	"fmt"
	"strconv"

	"github.com/bojanzelic/cloudflare-zero-trust-operator/api/v1alpha1"
	"github.com/bojanzelic/cloudflare-zero-trust-operator/internal/cfapi"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/api/equality"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
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
	if annotationPreventDestroy, ok := annotations[v1alpha1.AnnotationPreventDestroy]; ok {
		preventDestroy, _ = strconv.ParseBool(annotationPreventDestroy)
	}

	if preventDestroy && controllerutil.ContainsFinalizer(c, v1alpha1.FinalizerDeletion) {
		controllerutil.RemoveFinalizer(c, v1alpha1.FinalizerDeletion)
		if err := h.R.Update(ctx, c); err != nil {
			log.Error(err, "unable to remove finalizer")

			return errors.Wrap(err, "unable to remove finalizer")
		}
	} else if !controllerutil.ContainsFinalizer(c, v1alpha1.FinalizerDeletion) {
		controllerutil.AddFinalizer(c, v1alpha1.FinalizerDeletion)
		if err := h.R.Update(ctx, c); err != nil {
			log.Error(err, "unable to add finalizer")

			return errors.Wrap(err, "unable to add finalizer")
		}
	}

	return nil
}

func (h *ControllerHelper) StatusUpdate(ctx context.Context, obj CloudflareCR, f controllerutil.MutateFn) error {
	key := client.ObjectKeyFromObject(obj)
	existing := obj.DeepCopyObject().(CloudflareCR)

	if err := h.R.Get(ctx, key, obj); err != nil {
		return err
	}

	if equality.Semantic.DeepEqual(existing.GetStatus(), obj.GetStatus()) {
		fmt.Println("objects equal")
		return nil
	}

	fmt.Println("updating object")
	fmt.Println(obj)
	if err := h.R.Status().Update(ctx, obj); err != nil {
		return errors.Wrap(err, "update error")
	}

	return nil
}

func (h *ControllerHelper) Update(ctx context.Context, c client.Client, obj client.Object, f controllerutil.MutateFn) (controllerutil.OperationResult, error) {
	key := client.ObjectKeyFromObject(obj)
	if err := c.Get(ctx, key, obj); err != nil {
		if !apierrors.IsNotFound(err) {
			return controllerutil.OperationResultNone, err
		}
		if err := mutate(f, key, obj); err != nil {
			return controllerutil.OperationResultNone, err
		}
		if err := c.Create(ctx, obj); err != nil {
			return controllerutil.OperationResultNone, err
		}
		return controllerutil.OperationResultCreated, nil
	}

	existing := obj.DeepCopyObject()
	if err := mutate(f, key, obj); err != nil {
		return controllerutil.OperationResultNone, err
	}

	if equality.Semantic.DeepEqual(existing, obj) {
		return controllerutil.OperationResultNone, nil
	}

	if err := c.Update(ctx, obj); err != nil {
		return controllerutil.OperationResultNone, err
	}

	return controllerutil.OperationResultUpdated, nil
}

func mutate(f controllerutil.MutateFn, key client.ObjectKey, obj client.Object) error {
	if err := f(); err != nil {
		return err
	}
	if newKey := client.ObjectKeyFromObject(obj); key != newKey {
		return fmt.Errorf("MutateFn cannot mutate object name and/or object namespace")
	}

	return nil
}

//nolint:cyclop
func (h *ControllerHelper) ReconcileDeletion(ctx context.Context, api *cfapi.API, k8sCR CloudflareCR) (bool, error) {
	log := logger.FromContext(ctx).WithName("finalizerHelper::ReconcileDeletion").WithValues(
		"type", k8sCR.GetType(),
		"name", k8sCR.GetName(),
		"namespace", k8sCR.GetNamespace(),
	)

	// examine DeletionTimestamp to determine if object is under deletion
	if !k8sCR.UnderDeletion() {
		if err := h.EnsureFinalizer(ctx, k8sCR); err != nil {
			return false, errors.Wrap(err, "unable to reconcile finalizer")
		}

		return true, nil
	}

	// The object is being deleted
	if controllerutil.ContainsFinalizer(k8sCR, v1alpha1.FinalizerDeletion) {
		// our finalizer is present, so lets handle any external dependency
		if k8sCR.GetID() != "" {
			log.Info("will remove resource in Cloudflare")
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

			log.Info("resource removed in Cloudflare")
		}

		// remove our finalizer from the list and update it.
		controllerutil.RemoveFinalizer(k8sCR, v1alpha1.FinalizerDeletion)
		if err := h.R.Update(ctx, k8sCR); err != nil {
			log.Error(err, "unable to remove finalizer")

			return false, errors.Wrap(err, "unable to remove finalizer")
		}
	}

	// Stop reconciliation as the item is being deleted
	log.Info("destroyed successfully")

	return false, nil
}
