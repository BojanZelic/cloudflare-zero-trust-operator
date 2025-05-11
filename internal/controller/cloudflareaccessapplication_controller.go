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
	"time"

	v4alpha1 "github.com/bojanzelic/cloudflare-zero-trust-operator/api/v4alpha1"
	"github.com/bojanzelic/cloudflare-zero-trust-operator/internal/cfapi"
	"github.com/bojanzelic/cloudflare-zero-trust-operator/internal/cfcompare"
	"github.com/bojanzelic/cloudflare-zero-trust-operator/internal/config"
	"github.com/bojanzelic/cloudflare-zero-trust-operator/internal/ctrlhelper"
	cloudflare "github.com/cloudflare/cloudflare-go/v4"
	"github.com/cloudflare/cloudflare-go/v4/zero_trust"
	"github.com/pkg/errors"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logger "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// CloudflareAccessApplicationReconciler reconciles a CloudflareAccessApplication object.
type CloudflareAccessApplicationReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Helper *ctrlhelper.ControllerHelper
}

const (
	// statusAvailable represents the status of the Cloudflare App.
	statusAvailable = "Available"
	// statusDegraded represents the status used when the custom resource is deleted and the finalizer operations are must to occur.
	statusDegraded = "Degraded"
)

// +kubebuilder:rbac:groups=cloudflare.zelic.io,resources=cloudflareaccessapplications,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cloudflare.zelic.io,resources=cloudflareaccessapplications/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=cloudflare.zelic.io,resources=cloudflareaccessapplications/finalizers,verbs=update

//nolint:maintidx,cyclop,gocognit,gocyclo,varnamelen
func (r *CloudflareAccessApplicationReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logger.FromContext(ctx).WithName("CloudflareAccessApplicationController::Reconcile")
	app := &v4alpha1.CloudflareAccessApplication{}
	var err error

	//
	// Try to get AccessApplication CRD Manifest
	//

	//
	if err = r.Get(ctx, req.NamespacedName, app); err != nil {
		// Not found ? might have been deleted; skip Reconciliation
		if k8serrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}

		// Else, return with failure
		log.Error(err, "Failed to get CloudflareAccessApplication")
		return ctrl.Result{}, errors.Wrap(err, "Failed to get CloudflareAccessApplication")
	}

	//
	// Ensure all PolicyRefs underlying [CloudflareAccessReusablePolicy] references exist and are available,
	// then store their CF IDs
	//

	//
	policyRefsNS, err := app.Spec.GetNamespacedPolicyRefs(req.Namespace)
	if err != nil {
		return ctrl.Result{}, errors.Wrap(err, "Failed to determine policy references")
	}

	//
	var rp v4alpha1.CloudflareAccessReusablePolicy //nolint:varnamelen
	orderedPolicyIds := []string{}
	for _, policyRefNS := range policyRefsNS {
		err := r.Get(ctx, policyRefNS, &rp)
		if err != nil {
			return ctrl.Result{}, errors.Wrap(err, "Referenced Policy Ref does not correspond to an existing CloudflareAccessReusablePolicy")
		}

		ready := false
		for _, cond := range rp.Status.Conditions {
			if cond.Type == statusAvailable && cond.Status == metav1.ConditionTrue {
				ready = true
				break
			}
		}
		if !ready {
			log.Info("Referenced CloudflareAccessReusablePolicy not available yet, requeuing")
			return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
		}

		// if ready, we know for sure that AccessReusablePolicyID exists, so we extract it
		orderedPolicyIds = append(orderedPolicyIds, rp.Status.AccessReusablePolicyID)
	}

	app.Status.ReusablePolicyIDs = orderedPolicyIds
	if err := r.Client.Status().Update(ctx, app); err != nil {
		return ctrl.Result{}, errors.Wrap(err, "Failed to update CloudflareAccessApplication status")
	}

	//
	// Setup access to CF API
	//

	// Gather credentials to connect to Cloudflare's API
	cfConfig := config.ParseCloudflareConfig(app)
	validConfig, err := cfConfig.IsValid()
	if !validConfig {
		return ctrl.Result{}, errors.Wrap(err, "invalid config")
	}

	// Initialize Cloudflare's API wrapper
	api := cfapi.New(cfConfig.APIToken, cfConfig.APIKey, cfConfig.APIEmail, cfConfig.AccountID)

	//
	// Proceed marked-as pending operations of manifest (if any)
	//

	// Attempt pending deletions on CRD Manifest
	continueReconcilliation, err := r.Helper.ReconcileDeletion(ctx, api, app)
	if !continueReconcilliation || err != nil {
		if err != nil {
			log.Error(err, "unable to reconcile deletion")
		}
		return ctrl.Result{}, errors.Wrap(err, "unable to reconcile deletion")
	}

	//
	// May mark manifest status state
	//

	// Try to setup "Conditions" field on CRD Manifest associated status
	_, err = controllerutil.CreateOrPatch(ctx, r.Client, app, func() error {
		// if already has "Conditions", nothing to do
		if len(app.Status.Conditions) > 0 {
			return nil
		}

		// else, define it as reconciling
		meta.SetStatusCondition(&app.Status.Conditions,
			metav1.Condition{
				Type:    statusAvailable,
				Status:  metav1.ConditionUnknown,
				Reason:  "Reconciling",
				Message: "CloudflareAccessApplication is reconciling",
			},
		)
		return nil
	})
	if err != nil {
		return ctrl.Result{}, errors.Wrap(err, "Failed to update CloudflareAccessApplication status")
	}

	var cfAccessApp *zero_trust.AccessApplicationGetResponse

	//
	// Find CloudFlare AccessApplication UUID associated to this manifest (if missing)
	//

	if app.Status.AccessApplicationID == "" { //nolint
		switch app.Spec.Type {
		case "self_hosted":
			{
				cfAccessApp, err = api.FindAccessApplicationByDomain(ctx, app.Spec.Domain)
				if cfAccessApp == nil || err != nil {
					return ctrl.Result{}, errors.Wrap(err, "error querying application app from cloudflare")
				}
			}
		case "warp":
		case "app_launcher":
			{
				cfAccessApp, err = api.FindFirstAccessApplicationOfType(ctx, app.Spec.Type)
				if cfAccessApp == nil || err != nil {
					return ctrl.Result{}, errors.Wrapf(
						err,
						"error querying unique '%s' application from cloudflare. "+
							"Make sure you activated the associated feature in your cloudflare's dashboard !",
						app.Spec.Type,
					)
				}
			}
		default:
			{
				return ctrl.Result{}, errors.Errorf("Unhandled application type '%s'. Contact the developers.", app.Spec.Type)
			}
		}

		//
		if err = r.ReconcileStatus(ctx, cfAccessApp, app); err != nil {
			return ctrl.Result{}, errors.Wrap(err, "issue updating status")
		}
	}

	//
	// Try to get existing application from CloudFlare API (if existing)
	//

	if cfAccessApp == nil {
		cfAccessApp, err = api.AccessApplication(ctx, app.Status.AccessApplicationID)
		if err != nil {
			var cfErr *cloudflare.Error
			isNotFound := errors.As(err, &cfErr) && cfErr.StatusCode == 404

			// do not allow to continue if anything other than not found
			if !isNotFound {
				return ctrl.Result{}, errors.Wrap(err, "unable to get access application")
			}

			// well, Application ID we had do not exist anymore; lets recreate the app in CF
			log.Info("access application ID linked to manifest not found - recreating remote resource...", "accessApplicationID", app.Status.AccessApplicationID)
			app.Status.AccessApplicationID = ""
		}
	}

	//
	// May create / recreate / update Access Application on CloudFlare API
	//

	if cfAccessApp == nil {
		log.Info("app is missing - creating...", "name", app.Spec.Name, "domain", app.Spec.Domain)
		cfAccessApp, err = api.CreateAccessApplication(ctx, app)
		if err != nil {
			return ctrl.Result{}, errors.Wrap(err, "unable to create access application")
		}

		// update status
		if err = r.ReconcileStatus(ctx, cfAccessApp, app); err != nil {
			return ctrl.Result{}, errors.Wrap(err, "issue updating status")
		}
	} else if needsUpdate := !cfcompare.AreAccessApplicationsEquivalent(cfAccessApp, app); needsUpdate {
		log.Info("app has changed - updating...", "name", app.Spec.Name, "domain", app.Spec.Domain)
		cfAccessApp, err = api.UpdateAccessApplication(ctx, app)
		if err != nil {
			return ctrl.Result{}, errors.Wrap(err, "unable to update access group")
		}

		if err = r.ReconcileStatus(ctx, cfAccessApp, app); err != nil {
			return ctrl.Result{}, errors.Wrap(err, "issue updating status")
		}
	}

	//
	// All set, now mark ressource as available
	//

	if _, err = controllerutil.CreateOrPatch(ctx, r.Client, app, func() error {
		meta.SetStatusCondition(&app.Status.Conditions,
			metav1.Condition{
				Type:    statusAvailable,
				Status:  metav1.ConditionTrue,
				Reason:  "Reconcilied",
				Message: "App Reconciled Successfully",
			},
		)

		return nil
	}); err != nil {
		return ctrl.Result{}, errors.Wrap(err, "Failed to update CloudflareAccessApplication status")
	}

	return ctrl.Result{}, nil
}

//nolint:dupl
func (r *CloudflareAccessApplicationReconciler) ReconcileStatus(ctx context.Context, cfApp *zero_trust.AccessApplicationGetResponse, k8sApp *v4alpha1.CloudflareAccessApplication) error {
	if k8sApp.Status.AccessApplicationID != "" {
		return nil
	}
	if cfApp == nil {
		return nil
	}

	k8sApp.Status.AccessApplicationID = cfApp.ID
	k8sApp.Status.CreatedAt = metav1.NewTime(cfApp.CreatedAt)
	k8sApp.Status.UpdatedAt = metav1.NewTime(cfApp.UpdatedAt)

	if err := r.Client.Status().Update(ctx, k8sApp); err != nil {
		return errors.Wrap(err, "Failed to update CloudflareAccessApplication status")
	}
	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *CloudflareAccessApplicationReconciler) SetupWithManager(mgr ctrl.Manager) error {
	//nolint:wrapcheck
	return ctrl.NewControllerManagedBy(mgr).
		For(&v4alpha1.CloudflareAccessApplication{}, builder.WithPredicates(predicate.GenerationChangedPredicate{})).
		Complete(r)
}
