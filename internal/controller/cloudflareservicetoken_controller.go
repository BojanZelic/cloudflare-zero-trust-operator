/*
Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"

	"github.com/Southclaws/fault"
	"github.com/Southclaws/fault/fctx"
	"github.com/Southclaws/fault/fmsg"
	v4alpha1 "github.com/bojanzelic/cloudflare-zero-trust-operator/api/v4alpha1"
	"github.com/bojanzelic/cloudflare-zero-trust-operator/internal/cfapi"
	"github.com/bojanzelic/cloudflare-zero-trust-operator/internal/cftypes"
	"github.com/bojanzelic/cloudflare-zero-trust-operator/internal/config"
	"github.com/bojanzelic/cloudflare-zero-trust-operator/internal/ctrlhelper"
	"github.com/bojanzelic/cloudflare-zero-trust-operator/internal/meta"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav2 "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// CloudflareServiceTokenReconciler reconciles a CloudflareServiceToken object.
type CloudflareServiceTokenReconciler struct {
	CloudflareAccessReconciler
	client.Client
	Scheme *runtime.Scheme
	Helper *ctrlhelper.ControllerHelper

	// Mainly used for debug / tests purposes. Should not be instantiated in production run.
	OptionalTracer *cfapi.CloudflareResourceCreationTracer
}

// +kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cloudflare.zelic.io,resources=cloudflareservicetokens,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cloudflare.zelic.io,resources=cloudflareservicetokens/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=cloudflare.zelic.io,resources=cloudflareservicetokens/finalizers,verbs=update

func (r *CloudflareServiceTokenReconciler) GetReconcilierLogger(ctx context.Context) logr.Logger {
	return ctrl.LoggerFrom(ctx).WithName("CloudflareServiceTokenController::Reconcile")
}

//nolint:gocognit,cyclop,gocyclo,maintidx
func (r *CloudflareServiceTokenReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	var err error
	var api *cfapi.API

	log := r.GetReconcilierLogger(ctx)

	//
	// Find the service token resource
	//

	serviceToken := &v4alpha1.CloudflareServiceToken{}
	err = r.Get(ctx, req.NamespacedName, serviceToken)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			// will stop
			return ctrl.Result{}, nil
		}

		// will retry immediately
		return ctrl.Result{}, fault.Wrap(err,
			fmsg.With("Failed to get CloudflareServiceToken"),
			fctx.With(ctx, "serviceTokenName", req.Name),
		)
	}

	//
	// Configure CF API access
	//

	cfConfig := config.ParseCloudflareConfig(serviceToken)
	validConfig, err := cfConfig.IsValid()
	if !validConfig {
		// will retry immediately
		return ctrl.Result{}, fault.Wrap(err, fmsg.With("invalid config"))
	}
	api = cfapi.FromConfig(ctx, cfConfig, r.OptionalTracer)

	//
	//
	//

	continueReconcilliation, err := r.Helper.ReconcileDeletion(ctx, api, serviceToken)
	if !continueReconcilliation || err != nil {
		// will retry immediately
		return ctrl.Result{}, fault.Wrap(err, fmsg.With("unable to reconcile deletion for service token"))
	}

	//
	//
	//

	newCond := metav1.Condition{
		Type:    StatusAvailable,
		Status:  metav1.ConditionUnknown,
		Reason:  "Reconciling",
		Message: "ServiceToken is reconciling",
	}
	_, err = controllerutil.CreateOrPatch(ctx, r.Client, serviceToken, func() error {
		metav2.SetStatusCondition(&serviceToken.Status.Conditions, newCond)
		return nil
	})
	if err != nil {
		// will retry immediately
		return ctrl.Result{}, fault.Wrap(err, fmsg.With("Failed to update CloudflareServiceToken status"))
	} else {
		log.V(1).Info("Status persisted",
			"type", newCond.Type,
			"to", newCond.Status,
		)
	}

	//
	// Try to find an existing associated Secret
	//

	// this is used just for populating existingServiceToken
	associatedSecretList := &corev1.SecretList{}
	if err = r.List(ctx, associatedSecretList,
		client.MatchingLabels{meta.LabelOwnedBy: serviceToken.Name},
		client.InNamespace(serviceToken.Namespace),
	); err != nil {
		// will retry immediately
		return ctrl.Result{}, fault.Wrap(err, fmsg.With("unable to list secrets associated with a CloudflareServiceToken"))
	}

	// should not happen
	if len(associatedSecretList.Items) > 1 {
		log.Info("Found multiple secrets with the same owner label",
			"label", meta.LabelOwnedBy,
			"owner", serviceToken.Name,
		)
	}

	//
	associatedSecret := func() *corev1.Secret {
		if len(associatedSecretList.Items) > 0 {
			return &associatedSecretList.Items[0] // exists
		}
		return nil
	}()

	//
	// get CF Service Token From its associated secret (if possible)
	//

	var existingServiceToken *cftypes.ExtendedServiceToken
	if associatedSecret != nil {
		// CF UUID bound to Secret
		cfTokenID := string(associatedSecret.Data[associatedSecret.Annotations[meta.AnnotationTokenIDKey]])
		existingServiceToken, err = api.AccessServiceToken(ctx, cfTokenID)

		if err != nil {
			log.Info("Secret bound to CloudflareServiceToken does not refer to an existing Cloudflare Service Token. Recreating.")
		}
	}

	//
	// else, create it
	//

	if existingServiceToken == nil {
		existingServiceToken, err = api.CreateAccessServiceToken(ctx, serviceToken.ToExtendedToken())
		if err != nil {
			// will retry immediately
			return ctrl.Result{}, fault.Wrap(err, fmsg.With("unable to create access service token"))
		}

		log.Info("created access service token",
			"token_id", existingServiceToken.ID,
		)
	}

	//
	//
	//

	if associatedSecret != nil {
		if err = existingServiceToken.SetSecretValues(*associatedSecret); err != nil {
			// will retry immediately
			return ctrl.Result{}, fault.Wrap(err, fmsg.With("failed to set Secret values, associated to CloudflareServiceToken"))
		}
	}

	//
	//
	//

	// reconcile  associatedSecret
	wantedSecretMeta := metav1.ObjectMeta{
		Namespace: req.Namespace,
		Name:      req.Name,
	}
	if serviceToken.Spec.Template.Name != "" {
		wantedSecretMeta.Name = serviceToken.Spec.Template.Name
	}

	var secretToDelete *corev1.Secret
	// associatedSecret exists & was renamed; remove the old one
	if associatedSecret != nil && wantedSecretMeta.Name != associatedSecret.Name {
		// so, we need to remove the old one
		secretToDelete = associatedSecret
	}

	//
	//
	//

	associatedSecret = &corev1.Secret{
		ObjectMeta: wantedSecretMeta,
	}

	op, err := controllerutil.CreateOrUpdate(ctx, r.Client, associatedSecret, func() error { //nolint:varnamelen
		const serviceTokenID_label = "serviceTokenID"

		//
		//
		//
		secretAnnotations := map[string]string{
			meta.AnnotationClientIDKey:     serviceToken.Spec.Template.ClientIDKey,
			meta.AnnotationClientSecretKey: serviceToken.Spec.Template.ClientSecretKey,
			meta.AnnotationTokenIDKey:      serviceTokenID_label,
		}
		if serviceToken.Spec.Template.Annotations != nil {
			for annotationKey, annotationValue := range serviceToken.Spec.Template.Annotations {
				if _, exists := secretAnnotations[annotationKey]; !exists {
					secretAnnotations[annotationKey] = annotationValue
				}
			}
		}
		associatedSecret.SetAnnotations(secretAnnotations)

		//
		//
		//
		secretLabels := map[string]string{
			meta.LabelOwnedBy: serviceToken.Name,
		}
		if serviceToken.Spec.Template.Labels != nil {
			for labelKey, labelValue := range serviceToken.Spec.Template.Labels {
				if _, exists := secretLabels[labelKey]; !exists {
					secretLabels[labelKey] = labelValue
				}
			}
		}
		associatedSecret.SetLabels(secretLabels)

		//
		//
		//
		associatedSecret.Data = map[string][]byte{
			serviceToken.Spec.Template.ClientSecretKey: []byte(existingServiceToken.ClientSecret),
			serviceToken.Spec.Template.ClientIDKey:     []byte(existingServiceToken.ClientID),
			serviceTokenID_label:                       []byte(existingServiceToken.ID),
		}

		//
		//
		//

		if err = existingServiceToken.SetSecretValues(*associatedSecret); err != nil {
			return fault.Wrap(err, fmsg.With("unable to CreateOrUpdate Secret associated to CloudflareServiceToken"))
		}

		if err = ctrl.SetControllerReference(serviceToken, associatedSecret, r.Scheme); err != nil {
			return fault.Wrap(err, fmsg.With("unable to set Secret owner reference, associated to CloudflareServiceToken"))
		}

		return nil
	})

	if err != nil {
		// will retry immediately
		return ctrl.Result{}, fault.Wrap(err, fmsg.With("Failed to create/update Secret associated to CloudflareServiceToken"))
	}

	switch op {
	case controllerutil.OperationResultCreated:
		log.Info("created Secret associated to CloudflareServiceToken")
	case controllerutil.OperationResultUpdated:
		log.Info("updated Secret associated to CloudflareServiceToken")
	}

	//
	//
	//

	if secretToDelete != nil {
		if err = r.Delete(ctx, secretToDelete); err != nil {
			log.Error(err, "failed to remove old Secret associated to CloudflareServiceToken")
		} else {
			log.Info("removed old Secret associated to CloudflareServiceToken")
		}
	}

	//
	//
	//

	err = r.MayReconcileStatus(ctx, existingServiceToken, serviceToken)
	if err != nil {
		// will retry immediately
		return ctrl.Result{}, fault.Wrap(err, fmsg.With("unable to set status"))
	}

	newCond = metav1.Condition{
		Type:    StatusAvailable,
		Status:  metav1.ConditionTrue,
		Reason:  "Reconcilied",
		Message: "CloudflareServiceToken Reconciled Successfully",
	}
	_, err = controllerutil.CreateOrPatch(ctx, r.Client, serviceToken, func() error {
		metav2.SetStatusCondition(&serviceToken.Status.Conditions, newCond)
		return nil
	})
	if err != nil {
		// will retry immediately
		return ctrl.Result{}, fault.Wrap(err, fmsg.With("Failed to update CloudflareServiceToken status"))
	} else {
		log.V(1).Info("Status persisted",
			"type", newCond.Type,
			"to", newCond.Status,
		)
	}

	//
	// All good !
	//

	log.Info("changes successfully acknoledged")

	// will stop normally
	return ctrl.Result{}, nil
}

func (r *CloudflareServiceTokenReconciler) MayReconcileStatus(
	ctx context.Context,
	cfToken *cftypes.ExtendedServiceToken,
	k8sToken *v4alpha1.CloudflareServiceToken,
) error {
	if cfToken == nil {
		return nil
	}

	token := k8sToken.DeepCopy()

	_, err := controllerutil.CreateOrPatch(ctx, r.Client, token, func() error {
		token.Status.AccessServiceTokenID = cfToken.ID
		token.Status.CreatedAt = metav1.NewTime(cfToken.CreatedAt)
		token.Status.UpdatedAt = metav1.NewTime(cfToken.UpdatedAt)
		token.Status.ExpiresAt = metav1.NewTime(cfToken.ExpiresAt)
		token.Status.SecretRef = v4alpha1.SecretRef{
			LocalObjectReference: corev1.LocalObjectReference{
				Name: cfToken.K8sSecretRef.SecretName,
			},
			ClientSecretKey: cfToken.K8sSecretRef.ClientSecretKey,
			ClientIDKey:     cfToken.K8sSecretRef.ClientIDKey,
		}

		return nil
	})
	if err != nil {
		return fault.Wrap(err, fmsg.With("Failed to update CloudflareServiceToken"))
	} else {
		r.GetReconcilierLogger(ctx).V(1).Info("UUID persisted in status",
			"UUID", token.GetCloudflareUUID(),
		)
	}

	// CreateOrPatch re-fetches the object from k8s which removes any changes we've made that override them
	// so thats why we re-apply these settings again on the original object;
	k8sToken.Status = token.Status

	return nil
}

//
//
//

// SetupWithManager sets up the controller with the Manager.
func (r *CloudflareServiceTokenReconciler) SetupWithManager(mgr ctrl.Manager, override reconcile.Reconciler) error {
	if override == nil {
		override = r
	}

	//nolint:wrapcheck
	return ctrl.NewControllerManagedBy(mgr).
		For(&v4alpha1.CloudflareServiceToken{}, builder.WithPredicates(predicate.GenerationChangedPredicate{})).
		Owns(&corev1.Secret{}).
		WithOptions(controller.Options{
			RateLimiter: ZTOTypedControllerRateLimiter[reconcile.Request](),
		}).
		Complete(override)
}
