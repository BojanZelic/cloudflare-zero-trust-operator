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
	"github.com/bojanzelic/cloudflare-zero-trust-operator/internal/cfcompare"
	"github.com/bojanzelic/cloudflare-zero-trust-operator/internal/config"
	"github.com/bojanzelic/cloudflare-zero-trust-operator/internal/ctrlhelper"
	"github.com/cloudflare/cloudflare-go/v4/zero_trust"
	"github.com/go-logr/logr"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
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

// CloudflareAccessReusablePolicyReconciler reconciles a CloudflareAccessReusablePolicy object.
type CloudflareAccessReusablePolicyReconciler struct {
	CloudflareAccessReconciler
	client.Client
	Scheme *runtime.Scheme
	Helper *ctrlhelper.ControllerHelper

	// Mainly used for debug / tests purposes. Should not be instantiated in production run.
	OptionalTracer *cfapi.CloudflareResourceCreationTracer
}

// +kubebuilder:rbac:groups=cloudflare.zelic.io,resources=cloudflareaccessreusablepolicies,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cloudflare.zelic.io,resources=cloudflareaccessreusablepolicies/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=cloudflare.zelic.io,resources=cloudflareaccessreusablepolicies/finalizers,verbs=update

func (r *CloudflareAccessReusablePolicyReconciler) GetReconcilierLogger(ctx context.Context) logr.Logger {
	return ctrl.LoggerFrom(ctx).WithName("CloudflareAccessReusablePolicyController::Reconcile")
}

//nolint:cyclop,gocognit,maintidx
func (r *CloudflareAccessReusablePolicyReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.GetReconcilierLogger(ctx)
	reusablePolicy := &v4alpha1.CloudflareAccessReusablePolicy{}
	var err error

	//
	// Try to get AccessReusablePolicy CRD Manifest
	//

	//
	if err = r.Get(ctx, req.NamespacedName, reusablePolicy); err != nil {
		// Not found ? might have been deleted; skip Reconciliation
		if k8serrors.IsNotFound(err) {
			return ctrl.Result{}, nil // will stop
		}

		// Else, return with failure
		return ctrl.Result{}, fault.Wrap(err,
			fmsg.With("Failed to get CloudflareAccessReusablePolicy"),
			fctx.With(ctx, "policyName", req.Name),
		) // will retry immediately
	}

	//
	// Populate UUIDs
	//

	//
	popRes, populatedCount, err := r.Helper.PopulateWithCloudflareUUIDs(ctx, req.Namespace, &log, reusablePolicy)

	// if any result returned, return it to reconcilier along w/ err (if any)
	if popRes != nil {
		// add a wrap for error, but no error ever should be passed here
		return *popRes, fault.Wrap(err, fmsg.With("An unexpected error has been provided. Contact the developers."))
	} else if err != nil {
		//
		log.Info("failed to update access reusable policy's referenced CloudFlare UUIDs")

		// patch with status updated
		newCond := metav1.Condition{
			Type:    StatusAvailable,
			Status:  metav1.ConditionFalse,
			Reason:  "InvalidReference",
			Message: err.Error(),
		}
		_, pErr := controllerutil.CreateOrPatch(ctx, r.Client, reusablePolicy, func() error {
			meta.SetStatusCondition(&reusablePolicy.Status.Conditions, newCond)
			return nil
		})
		if pErr != nil {
			// will retry immediately
			return ctrl.Result{}, fault.Wrap(pErr, fmsg.With("Failed to update CloudflareAccessReusablePolicy status, after a CF UUIDs population failure"))
		} else {
			log.V(1).Info("Status persisted",
				"type", newCond.Type,
				"to", newCond.Status,
			)
		}

		// will retry immediately
		return ctrl.Result{}, fault.Wrap(err, fmsg.With("Failed to populate CF UUIDs"))
	} else if populatedCount > 0 {

		//
		// Record populated values
		//

		if err = r.Client.Status().Update(ctx, reusablePolicy); err != nil {
			// will retry immediately
			return ctrl.Result{}, fault.Wrap(err, fmsg.With("Failed to update CloudflareAccessApplication status"))
		} else {
			log.V(1).Info("Persisted Populated UUIDs", "populatedCount", populatedCount)
		}
	}

	//
	// Setup access to CF API
	//

	// Gather credentials to connect to Cloudflare's API
	cfConfig := config.ParseCloudflareConfig(reusablePolicy)
	validConfig, err := cfConfig.IsValid()
	if !validConfig {
		// will retry immediately
		return ctrl.Result{}, fault.Wrap(err, fmsg.With("invalid config"))
	}

	// Initialize Cloudflare's API wrapper
	api := cfapi.FromConfig(ctx, cfConfig, r.OptionalTracer)

	//
	// Proceed marked-as pending operations of manifest (if any)
	//

	// Attempt pending deletions on CRD Manifest
	continueReconcilliation, err := r.Helper.ReconcileDeletion(ctx, api, reusablePolicy)
	if !continueReconcilliation || err != nil {
		// will retry immediately
		return ctrl.Result{}, fault.Wrap(err, fmsg.With("unable to reconcile deletion for access reusable policy"))
	}

	//
	// May mark manifest status state
	//

	// Try to setup "Conditions" field on CRD Manifest associated status
	newCond := metav1.Condition{
		Type:    StatusAvailable,
		Status:  metav1.ConditionUnknown,
		Reason:  "Reconciling",
		Message: "CloudflareAccessReusablePolicy is reconciling",
	}
	_, err = controllerutil.CreateOrPatch(ctx, r.Client, reusablePolicy, func() error {
		meta.SetStatusCondition(&reusablePolicy.Status.Conditions, newCond)
		return nil
	})
	if err != nil {
		// will retry immediately
		return ctrl.Result{}, fault.Wrap(err, fmsg.With("Failed to update CloudflareAccessReusablePolicy status"))
	} else {
		log.V(1).Info("Status persisted",
			"type", newCond.Type,
			"to", newCond.Status,
		)
	}

	//
	// Find CloudFlare Access Reusable Policy depending on presence of CF UUID bound to resource
	//

	var cfAccessReusablePolicy *zero_trust.AccessPolicyGetResponse

	//
	if reusablePolicy.GetCloudflareUUID() == "" {

		//
		// Has no UUID
		//

		//
		// Unlike AccessApplications with domain and AccessGroup with name, since we have no equivalent of UUID to match from existing,
		// we'll just create anew then !
		//

	} else {

		//
		// Already has a UUID
		//

		//
		cfAccessReusablePolicy, err = api.AccessReusablePolicy(ctx, reusablePolicy.GetCloudflareUUID())

		//
		if err != nil {
			// do not allow to continue if anything other than not found
			if !api.Is404(err) {
				// will retry immediately
				return ctrl.Result{}, fault.Wrap(err, fmsg.With("unable to get access reusable policy"))
			}

			// well, Application ID we had do not exist anymore; lets recreate the app in CF
			log.Info("access reusable policy ID linked to manifest not found - recreating remote resource...",
				"accessReusablePolicyID", reusablePolicy.GetCloudflareUUID(),
			)

			// reset UUID so we can reconcile later on
			reusablePolicy.Status.AccessReusablePolicyID = ""
		}

	}

	//
	// May create / recreate / update Access Group on CloudFlare API
	//

	if cfAccessReusablePolicy == nil {

		//
		// no ressource found, create it with API
		//

		log.Info("reusable policy is missing - creating...",
			"name", reusablePolicy.Spec.Name,
		)

		//
		cfAccessReusablePolicy, err = api.CreateAccessReusablePolicy(ctx, reusablePolicy)
		if err != nil {
			// will retry immediately
			return ctrl.Result{}, fault.Wrap(err, fmsg.With("unable to create reusable policy"))
		}

		log.Info("reusable policy successfully created !")

		//
		err = r.MayReconcileStatus(ctx, cfAccessReusablePolicy, reusablePolicy)
		if err != nil {
			// will retry immediately
			return ctrl.Result{}, fault.Wrap(err, fmsg.With("unable to set reusable policy status"))
		}

	} else if mustUpdate := !cfcompare.AreAccessReusablePoliciesEquivalent(cfAccessReusablePolicy, reusablePolicy); mustUpdate {

		//
		// diff found between fetched CF resource and definition
		//

		log.Info("reusable policy has changed - updating...",
			"name", reusablePolicy.Spec.Name,
		)

		//
		err = api.UpdateAccessReusablePolicy(ctx, reusablePolicy)
		if err != nil {
			// will retry immediately
			return ctrl.Result{}, fault.Wrap(err, fmsg.With("unable to update reusable policies"))
		}

		log.Info("reusable policy successfully updated !")

		//
		err = r.MayReconcileStatus(ctx, cfAccessReusablePolicy, reusablePolicy)
		if err != nil {
			// will retry immediately
			return ctrl.Result{}, fault.Wrap(err, fmsg.With("unable to set reusable policy status"))
		}
	}

	//
	// All set, now mark ressource as available
	//

	newCond = metav1.Condition{
		Type:    StatusAvailable,
		Status:  metav1.ConditionTrue,
		Reason:  "Reconcilied",
		Message: "App Reconciled Successfully",
	}
	_, err = controllerutil.CreateOrPatch(ctx, r.Client, reusablePolicy, func() error {
		meta.SetStatusCondition(&reusablePolicy.Status.Conditions, newCond)
		return nil
	})
	if err != nil {
		// will retry immediately
		return ctrl.Result{}, fault.Wrap(err, fmsg.With("Failed to update CloudflareAccessReusablePolicy status"))
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

//nolint:dupl
func (r *CloudflareAccessReusablePolicyReconciler) MayReconcileStatus(
	ctx context.Context,
	cfReusablPolicy *zero_trust.AccessPolicyGetResponse,
	k8sReusablePolicy *v4alpha1.CloudflareAccessReusablePolicy,
) error {
	if k8sReusablePolicy.GetCloudflareUUID() != "" {
		return nil
	}

	if cfReusablPolicy == nil {
		return nil
	}

	reusablePolicy := k8sReusablePolicy.DeepCopy()

	_, err := controllerutil.CreateOrPatch(ctx, r.Client, reusablePolicy, func() error {
		reusablePolicy.Status.AccessReusablePolicyID = cfReusablPolicy.ID
		reusablePolicy.Status.CreatedAt = metav1.NewTime(cfReusablPolicy.CreatedAt)
		reusablePolicy.Status.UpdatedAt = metav1.NewTime(cfReusablPolicy.UpdatedAt)

		return nil
	})

	// CreateOrPatch re-fetches the object from k8s which removes any changes we've made that override them
	// so thats why we re-apply these settings again on the original object;
	k8sReusablePolicy.Status = reusablePolicy.Status

	if err != nil {
		return fault.Wrap(err, fmsg.With("Failed to update CloudflareAccessReusablePolicy status"))
	} else {
		r.GetReconcilierLogger(ctx).V(1).Info("UUID persisted in status",
			"UUID", reusablePolicy.GetCloudflareUUID(),
		)
	}

	return nil
}

//
//
//

// SetupWithManager sets up the controller with the Manager.
func (r *CloudflareAccessReusablePolicyReconciler) SetupWithManager(mgr ctrl.Manager, override reconcile.Reconciler) error {
	if override == nil {
		override = r
	}

	//nolint:wrapcheck
	return ctrl.NewControllerManagedBy(mgr).
		For(&v4alpha1.CloudflareAccessReusablePolicy{}, builder.WithPredicates(predicate.GenerationChangedPredicate{})).
		WithOptions(controller.Options{
			RateLimiter: ZTOTypedControllerRateLimiter[reconcile.Request](),
		}).
		Complete(override)
}
