/*
Copyright 2022.

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

	v4alpha1 "github.com/bojanzelic/cloudflare-zero-trust-operator/api/v4alpha1"
	"github.com/bojanzelic/cloudflare-zero-trust-operator/internal/cfapi"
	"github.com/bojanzelic/cloudflare-zero-trust-operator/internal/cfcollections"
	"github.com/bojanzelic/cloudflare-zero-trust-operator/internal/config"
	"github.com/bojanzelic/cloudflare-zero-trust-operator/internal/ctrlhelper"
	"github.com/bojanzelic/cloudflare-zero-trust-operator/internal/services"
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
	// statusDegrated represents the status used when the custom resource is deleted and the finalizer operations are must to occur.
	statusDegrated = "Degraded"
)

// +kubebuilder:rbac:groups=cloudflare.zelic.io,resources=cloudflareaccessapplications,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cloudflare.zelic.io,resources=cloudflareaccessapplications/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=cloudflare.zelic.io,resources=cloudflareaccessapplications/finalizers,verbs=update

//nolint:cyclop,gocognit
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
		//
		cfAccessApp, err = api.FindAccessApplicationByDomain(ctx, app.Spec.Domain)
		if cfAccessApp == nil || err != nil {
			return ctrl.Result{}, errors.Wrap(err, "error querying application app from cloudflare")
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
	} else if !cfcollections.AreAccessApplicationsEquivalent(cfAccessApp, app) {
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
	// Now, lets have a look at the associated policies
	//

	currentApplicationPolicies, err := api.AccessApplicationPolicies(ctx, app.Status.AccessApplicationID)
	currentApplicationPolicies.SortByPrecedence()
	if err != nil {
		return ctrl.Result{}, errors.Wrap(err, "unable get  access policies")
	}

	apService := &services.AccessApplicationPolicyRefMatcherService{
		Client: r.Client,
		Log:    log,
	}
	if err := apService.PopulateWithCloudflareUUIDs(ctx, app.Spec.Policies.ToGenericPolicyRuler()); err != nil {
		_, err = controllerutil.CreateOrPatch(ctx, r.Client, app, func() error {
			meta.SetStatusCondition(&app.Status.Conditions,
				metav1.Condition{
					Type:    statusDegrated,
					Status:  metav1.ConditionFalse,
					Reason:  "InvalidReference",
					Message: err.Error(),
				},
			)

			return nil
		})

		log.Info("failed to update  access policies")

		if err != nil {
			return ctrl.Result{}, errors.Wrap(err, "Failed to update CloudflareAccessApplication status")
		}

		// don't requeue
		return ctrl.Result{}, nil
	}
	expectedApplicationPolicies := app.Spec.Policies.ToCloudflare()
	expectedApplicationPolicies.SortByPrecedence()

	err = r.ReconcileApplicationPolicies(ctx, api, app, currentApplicationPolicies, expectedApplicationPolicies)
	if err != nil {
		return ctrl.Result{}, errors.Wrap(err, "unable get  access policies")
	}

	if _, err = controllerutil.CreateOrPatch(ctx, r.Client, app, func() error {
		meta.SetStatusCondition(&app.Status.Conditions,
			metav1.Condition{
				Type:    statusAvailable,
				Status:  metav1.ConditionTrue,
				Reason:  "Reconciling",
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

//nolint:gocognit,cyclop
func (r *CloudflareAccessApplicationReconciler) ReconcileApplicationPolicies(
	ctx context.Context, api *cfapi.API,
	app *v4alpha1.CloudflareAccessApplication,
	current, expected cfcollections.AccessApplicationPolicyCollection,
) error {
	log := logger.FromContext(ctx)

	for i := 0; i < len(current) || i < len(expected); i++ { //nolint:varnamelen
		var k8sPolicy *zero_trust.AccessApplicationPolicyListResponse
		var cfPolicy *zero_trust.AccessApplicationPolicyListResponse
		var err error
		var action string
		if i < len(current) {
			cfPolicy = &current[i]
		}
		if i < len(expected) {
			k8sPolicy = &expected[i]
		}

		if !cfcollections.AreK8SAccessPoliciesPresent(cfPolicy, k8sPolicy) {
			if cfPolicy == nil && k8sPolicy != nil {
				action = "create"
				log.Info("accesspolicy is missing - creating...", "policyName", k8sPolicy.Name, "domain", app.Spec.Domain)
				err = api.CreateAccessApplicationPolicies(ctx, app.Status.AccessApplicationID, *k8sPolicy)
			}
			if k8sPolicy == nil && cfPolicy != nil {
				action = "delete"
				log.Info("accesspolicy is removed - deleting...", "policyId", cfPolicy.ID, "policyName", cfPolicy.Name, "domain", app.Spec.Domain)
				err = api.DeleteAccessApplicationPolicy(ctx, app.Status.AccessApplicationID, cfPolicy.ID)
			}
			if cfPolicy != nil && k8sPolicy != nil {
				action = "update"
				k8sPolicy.ID = cfPolicy.ID
				log.Info("accesspolicy is changed - updating...", "policyId", cfPolicy.ID, "policyName", cfPolicy.Name, "domain", app.Spec.Domain)
				err = api.UpdateAccessApplicationPolicy(ctx, app.Status.AccessApplicationID, *k8sPolicy)
			}

			if err != nil {
				return errors.Wrapf(err, "Unable to %s access policy", action)
			}
		}
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
