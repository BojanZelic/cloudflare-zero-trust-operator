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

// CloudflareAccessGroupReconciler reconciles a CloudflareAccessGroup object.
type CloudflareAccessGroupReconciler struct {
	CloudflareAccessReconciler
	client.Client
	Scheme *runtime.Scheme
	Helper *ctrlhelper.ControllerHelper

	// Mainly used for debug / tests purposes. Should not be instantiated in production run.
	OptionalTracer *cfapi.CloudflareResourceCreationTracer
}

// +kubebuilder:rbac:groups=cloudflare.zelic.io,resources=cloudflareaccessgroups,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cloudflare.zelic.io,resources=cloudflareaccessgroups/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=cloudflare.zelic.io,resources=cloudflareaccessgroups/finalizers,verbs=update

func (r *CloudflareAccessGroupReconciler) GetReconcilierLogger(ctx context.Context) logr.Logger {
	return ctrl.LoggerFrom(ctx).WithName("CloudflareAccessGroupController::Reconcile")
}

//nolint:cyclop,gocognit,maintidx
func (r *CloudflareAccessGroupReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.GetReconcilierLogger(ctx)
	accessGroup := &v4alpha1.CloudflareAccessGroup{}
	var err error

	//
	// Try to get AccessGroup CRD Manifest
	//

	//
	if err = r.Get(ctx, req.NamespacedName, accessGroup); err != nil {
		// Not found ? might have been deleted; skip Reconciliation
		if k8serrors.IsNotFound(err) {
			return ctrl.Result{}, nil // will stop
		}

		// Else, return with failure
		return ctrl.Result{}, fault.Wrap(err,
			fmsg.With("Failed to get CloudflareAccessGroup"),
			fctx.With(ctx, "groupName", req.Name),
		) // will retry immediately
	}

	//
	// Populate UUIDs
	//

	//
	popRes, populatedCount, err := r.Helper.PopulateWithCloudflareUUIDs(ctx, req.Namespace, &log, accessGroup)

	// if any result returned, return it to reconcilier along w/ err (if any)
	if popRes != nil {
		// add a wrap for error, but no error ever should be passed here
		return *popRes, fault.Wrap(err, fmsg.With("An unexpected error has been provided. Contact the developers."))
	} else if err != nil {
		//
		log.Info("failed to update access group's referenced CloudFlare UUIDs")

		// patch with status updated
		newCond := metav1.Condition{
			Type:    StatusAvailable,
			Status:  metav1.ConditionFalse,
			Reason:  "InvalidReference",
			Message: err.Error(),
		}
		_, pErr := controllerutil.CreateOrPatch(ctx, r.Client, accessGroup, func() error {
			meta.SetStatusCondition(&accessGroup.Status.Conditions, newCond)
			return nil
		})
		if pErr != nil {
			// will retry immediately
			return ctrl.Result{}, fault.Wrap(pErr, fmsg.With("Failed to update CloudflareAccessGroup status, after a CF UUIDs population failure"))
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

		if err = r.Client.Status().Update(ctx, accessGroup); err != nil {
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
	cfConfig := config.ParseCloudflareConfig(accessGroup)
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
	continueReconcilliation, err := r.Helper.ReconcileDeletion(ctx, api, accessGroup)
	if !continueReconcilliation || err != nil {
		// will retry immediately
		return ctrl.Result{}, fault.Wrap(err, fmsg.With("unable to reconcile deletion for access group"))
	}

	//
	// May mark manifest status state
	//

	// Try to setup "Conditions" field on CRD Manifest associated status
	newCond := metav1.Condition{
		Type:    StatusAvailable,
		Status:  metav1.ConditionUnknown,
		Reason:  "Reconciling",
		Message: "CloudflareAccessGroup is reconciling",
	}
	_, err = controllerutil.CreateOrPatch(ctx, r.Client, accessGroup, func() error {
		meta.SetStatusCondition(&accessGroup.Status.Conditions, newCond)
		return nil
	})
	if err != nil {
		// will retry immediately
		return ctrl.Result{}, fault.Wrap(err, fmsg.With("Failed to update CloudflareAccessGroup status"))
	} else {
		log.V(1).Info("Status persisted",
			"type", newCond.Type,
			"to", newCond.Status,
		)
	}

	//
	// Find CloudFlare Access Group depending on presence of CF UUID bound to resource
	//

	var cfAccessGroup *zero_trust.AccessGroupGetResponse

	//
	if accessGroup.GetCloudflareUUID() == "" {

		//
		// Has no UUID
		//

		cfAccessGroup, err = api.AccessGroupByName(ctx, accessGroup.Spec.Name)
		if err != nil {
			// will retry immediately
			return ctrl.Result{}, fault.Wrap(err, fmsg.With("unable to get access group by name"))
		}

		// else...

		//
		// ...if not found, we'll create it later then !
		//

		// would do nothing if resource was not found above
		err = r.MayReconcileStatus(ctx, cfAccessGroup, accessGroup)
		if err != nil {
			// will retry immediately
			return ctrl.Result{}, fault.Wrap(err, fmsg.With("unable to update access groups"))
		}

	} else {

		//
		// Already has a UUID
		//

		//
		cfAccessGroup, err = api.AccessGroup(ctx, accessGroup.GetCloudflareUUID())

		//
		if err != nil {
			// do not allow to continue if anything other than not found
			if !api.Is404(err) {
				// will retry immediately
				return ctrl.Result{}, fault.Wrap(err, fmsg.With("unable to get access group"))
			}

			// well, Group ID we had do not exist anymore; lets recreate the app in CF
			log.Info("access group ID linked to manifest not found - recreating remote resource...",
				"accessGroupID", accessGroup.GetCloudflareUUID(),
			)

			// reset UUID so we can reconcile later on
			accessGroup.Status.AccessGroupID = ""
		}
	}

	//
	// May create / recreate / update Access Group on CloudFlare API
	//

	if cfAccessGroup == nil {

		//
		// no ressource found, create it with API
		//

		log.Info("group is missing - creating...",
			"name", accessGroup.Spec.Name,
		)

		//
		cfAccessGroup, err = api.CreateAccessGroup(ctx, accessGroup)
		if err != nil {
			// will retry immediately
			return ctrl.Result{}, fault.Wrap(err, fmsg.With("unable to create access group"))
		}

		log.Info("group successfully created !")

		// update status
		if err = r.MayReconcileStatus(ctx, cfAccessGroup, accessGroup); err != nil {
			// will retry immediately
			return ctrl.Result{}, fault.Wrap(err, fmsg.With("issue updating status"))
		}

	} else if needsUpdate := !cfcompare.AreAccessGroupsEquivalent(cfAccessGroup, accessGroup); needsUpdate {

		//
		// diff found between fetched CF resource and definition
		//

		log.Info("group has changed - updating...",
			"name", accessGroup.Spec.Name,
		)

		//
		err = api.UpdateAccessGroup(ctx, accessGroup)
		if err != nil {
			// will retry immediately
			return ctrl.Result{}, fault.Wrap(err, fmsg.With("unable to update access groups"))
		}

		log.Info("group successfully updated !")

		// update status
		if err = r.MayReconcileStatus(ctx, cfAccessGroup, accessGroup); err != nil {
			// will retry immediately
			return ctrl.Result{}, fault.Wrap(err, fmsg.With("issue updating status"))
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
	_, err = controllerutil.CreateOrPatch(ctx, r.Client, accessGroup, func() error {
		meta.SetStatusCondition(&accessGroup.Status.Conditions, newCond)
		return nil
	})
	if err != nil {
		// will retry immediately
		return ctrl.Result{}, fault.Wrap(err, fmsg.With("Failed to update CloudflareAccessGroup status"))
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
func (r *CloudflareAccessGroupReconciler) MayReconcileStatus(
	ctx context.Context,
	cfGroup *zero_trust.AccessGroupGetResponse,
	k8sGroup *v4alpha1.CloudflareAccessGroup,
) error {
	if k8sGroup.GetCloudflareUUID() != "" {
		return nil
	}

	if cfGroup == nil {
		return nil
	}

	group := k8sGroup.DeepCopy()

	_, err := controllerutil.CreateOrPatch(ctx, r.Client, group, func() error {
		group.Status.AccessGroupID = cfGroup.ID
		group.Status.CreatedAt = metav1.NewTime(cfGroup.CreatedAt)
		group.Status.UpdatedAt = metav1.NewTime(cfGroup.UpdatedAt)

		return nil
	})

	// CreateOrPatch re-fetches the object from k8s which removes any changes we've made that override them
	// so thats why we re-apply these settings again on the original object;
	k8sGroup.Status = group.Status

	if err != nil {
		return fault.Wrap(err, fmsg.With("Failed to update CloudflareAccessGroup status"))
	} else {
		r.GetReconcilierLogger(ctx).V(1).Info("UUID persisted in status",
			"UUID", group.GetCloudflareUUID(),
		)
	}

	return nil
}

//
//
//

// SetupWithManager sets up the controller with the Manager.
func (r *CloudflareAccessGroupReconciler) SetupWithManager(mgr ctrl.Manager, override reconcile.Reconciler) error {
	if override == nil {
		override = r
	}

	//nolint:wrapcheck
	return ctrl.NewControllerManagedBy(mgr).
		For(&v4alpha1.CloudflareAccessGroup{}, builder.WithPredicates(predicate.GenerationChangedPredicate{})).
		WithOptions(controller.Options{
			RateLimiter: ZTOTypedControllerRateLimiter[reconcile.Request](),
		}).
		Complete(override)
}
