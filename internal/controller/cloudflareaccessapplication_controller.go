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
	v1alpha1 "github.com/bojanzelic/cloudflare-zero-trust-operator/api/v1alpha1"
	"github.com/bojanzelic/cloudflare-zero-trust-operator/internal/cfapi"
	"github.com/bojanzelic/cloudflare-zero-trust-operator/internal/cfcollections"
	"github.com/bojanzelic/cloudflare-zero-trust-operator/internal/config"
	"github.com/bojanzelic/cloudflare-zero-trust-operator/internal/ctrlhelper"
	"github.com/bojanzelic/cloudflare-zero-trust-operator/internal/services"
	cloudflare "github.com/cloudflare/cloudflare-go"
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
	var err error
	var existingaccessApp *cloudflare.AccessApplication
	var api *cfapi.API

	log := logger.FromContext(ctx).WithName("CloudflareAccessApplicationController::Reconcile")

	app := &v1alpha1.CloudflareAccessApplication{}

	if err = r.Client.Get(ctx, req.NamespacedName, app); err != nil {
		if k8serrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}

		log.Error(err, "Failed to get CloudflareAccessApplication")

		return ctrl.Result{}, errors.Wrap(err, "Failed to get CloudflareAccessApplication")
	}

	cfConfig := config.ParseCloudflareConfig(app)
	validConfig, err := cfConfig.IsValid()
	if !validConfig {
		return ctrl.Result{}, errors.Wrap(err, "invalid config")
	}

	api, err = cfapi.New(cfConfig.APIToken, cfConfig.APIKey, cfConfig.APIEmail, cfConfig.AccountID)

	if err != nil {
		return ctrl.Result{}, errors.Wrap(err, "unable to initialize cloudflare object")
	}

	continueReconcilliation, err := r.Helper.ReconcileDeletion(ctx, api, app)
	if !continueReconcilliation || err != nil {
		if err != nil {
			log.Error(err, "unable to reconcile deletion")
		}

		return ctrl.Result{}, errors.Wrap(err, "unable to reconcile deletion")
	}

	_, err = controllerutil.CreateOrPatch(ctx, r.Client, app, func() error {
		if len(app.Status.Conditions) == 0 {
			meta.SetStatusCondition(&app.Status.Conditions, metav1.Condition{Type: statusAvailable, Status: metav1.ConditionUnknown, Reason: "Reconciling", Message: "CloudflareAccessApplication is reconciling"})
		}

		return nil
	})

	if err != nil {
		return ctrl.Result{}, errors.Wrap(err, "Failed to update CloudflareAccessApplication status")
	}

	apService := &services.AccessPolicyService{
		Client: r.Client,
		Log:    log,
	}

	if app.Status.AccessApplicationID == "" { // nolint
		accessApp, err := api.FindAccessApplicationByDomain(ctx, app.Spec.Domain)
		if err != nil {
			return ctrl.Result{}, errors.Wrap(err, "error querying application app from cloudflare")
		}

		err = r.ReconcileStatus(ctx, accessApp, app)
		if err != nil {
			return ctrl.Result{}, errors.Wrap(err, "issue updating status")
		}
	} else {
		accessApp, err := api.AccessApplication(ctx, app.Status.AccessApplicationID)
		if err != nil {
			var apiErr *cloudflare.NotFoundError
			if errors.As(err, &apiErr) {
				log.Info("access application not found - recreating...", "accessApplicationID", app.Status.AccessApplicationID)
				err := r.ClearStatus(ctx, app)
				if err != nil {
					return ctrl.Result{}, errors.Wrap(err, "failed to clear access application status")
				}
			} else {
				return ctrl.Result{}, errors.Wrap(err, "unable to get access application")
			}
		} else {
			existingaccessApp = &accessApp
		}
	}

	if existingaccessApp == nil {
		newApp := app.ToCloudflare()

		log.Info("app is missing - creating...", "name", app.Spec.Name, "domain", app.Spec.Domain)
		accessapp, err := api.CreateAccessApplication(ctx, newApp)
		existingaccessApp = &accessapp
		if err != nil {
			return ctrl.Result{}, errors.Wrap(err, "unable to create access group")
		}

		if err = r.ReconcileStatus(ctx, &accessapp, app); err != nil {
			return ctrl.Result{}, errors.Wrap(err, "issue updating status")
		}
	}

	if !cfcollections.AccessAppEqual(*existingaccessApp, app.ToCloudflare()) {
		log.Info("app has changed - updating...", "name", app.Spec.Name, "domain", app.Spec.Domain)
		accessapp, err := api.UpdateAccessApplication(ctx, app.ToCloudflare())
		if err != nil {
			return ctrl.Result{}, errors.Wrap(err, "unable to update access group")
		}

		err = r.ReconcileStatus(ctx, &accessapp, app)
		if err != nil {
			return ctrl.Result{}, errors.Wrap(err, "issue updating status")
		}
	}

	currentPolicies, err := api.AccessPolicies(ctx, app.Status.AccessApplicationID)
	currentPolicies.SortByPrecidence()
	if err != nil {
		return ctrl.Result{}, errors.Wrap(err, "unable get access policies")
	}

	if err := apService.PopulateAccessPolicyReferences(ctx, services.ToAccessPolicyList(app.Spec.Policies)); err != nil {
		_, err = controllerutil.CreateOrPatch(ctx, r.Client, app, func() error {
			meta.SetStatusCondition(&app.Status.Conditions, metav1.Condition{Type: statusDegrated, Status: metav1.ConditionFalse, Reason: "InvalidReference", Message: err.Error()})

			return nil
		})

		log.Info("failed to update access policies")

		if err != nil {
			return ctrl.Result{}, errors.Wrap(err, "Failed to update CloudflareAccessApplication status")
		}

		// don't requeue
		return ctrl.Result{}, nil
	}
	expectedPolicies := app.Spec.Policies.ToCloudflare()
	expectedPolicies.SortByPrecidence()

	err = r.ReconcilePolicies(ctx, api, app, currentPolicies, expectedPolicies)
	if err != nil {
		return ctrl.Result{}, errors.Wrap(err, "unable get access policies")
	}

	if _, err = controllerutil.CreateOrPatch(ctx, r.Client, app, func() error {
		meta.SetStatusCondition(&app.Status.Conditions, metav1.Condition{Type: statusAvailable, Status: metav1.ConditionTrue, Reason: "Reconciling", Message: "App Reconciled Successfully"})

		return nil
	}); err != nil {
		return ctrl.Result{}, errors.Wrap(err, "Failed to update CloudflareAccessApplication status")
	}

	return ctrl.Result{}, nil
}

type statusReconcileFunction func(cfApp *cloudflare.AccessApplication, k8sApp *v1alpha1.CloudflareAccessApplication) error

// nolint:dupl
func (r *CloudflareAccessApplicationReconciler) reconcileStatus(ctx context.Context, cfApp *cloudflare.AccessApplication, k8sApp *v1alpha1.CloudflareAccessApplication, reconcileFunction statusReconcileFunction) error {
	app := k8sApp.DeepCopy()

	if _, err := controllerutil.CreateOrPatch(ctx, r.Client, app, func() error {
		return reconcileFunction(cfApp, app)
	}); err != nil {
		return errors.Wrap(err, "Failed to reconcile CloudflareAccessApplication status")
	}

	// CreateOrPatch re-fetches the object from k8s which removes any changes we've made that override them
	// so that's why we re-apply these settings again on the original object;
	k8sApp.Status = app.Status

	return nil
}

// nolint:dupl
func (r *CloudflareAccessApplicationReconciler) ClearStatus(ctx context.Context, k8sApp *v1alpha1.CloudflareAccessApplication) error {
	if k8sApp.Status.AccessApplicationID == "" {
		return nil
	}

	return r.reconcileStatus(ctx, nil, k8sApp, func(_ *cloudflare.AccessApplication, app *v1alpha1.CloudflareAccessApplication) error {
		app.Status.AccessApplicationID = ""

		return nil
	})
}

// nolint:dupl
func (r *CloudflareAccessApplicationReconciler) ReconcileStatus(ctx context.Context, cfApp *cloudflare.AccessApplication, k8sApp *v1alpha1.CloudflareAccessApplication) error {
	if k8sApp.Status.AccessApplicationID != "" {
		return nil
	}

	if cfApp == nil {
		return nil
	}

	return r.reconcileStatus(ctx, cfApp, k8sApp, func(cfApp *cloudflare.AccessApplication, app *v1alpha1.CloudflareAccessApplication) error {
		app.Status.AccessApplicationID = cfApp.ID
		app.Status.CreatedAt = metav1.NewTime(*cfApp.CreatedAt)
		app.Status.UpdatedAt = metav1.NewTime(*cfApp.UpdatedAt)

		return nil
	})
}

//nolint:gocognit,cyclop
func (r *CloudflareAccessApplicationReconciler) ReconcilePolicies(ctx context.Context, api *cfapi.API, app *v1alpha1.CloudflareAccessApplication, current, expected cfcollections.AccessPolicyCollection) error {
	log := logger.FromContext(ctx)

	for i := 0; i < len(current) || i < len(expected); i++ { //nolint:varnamelen
		var k8sPolicy *cloudflare.AccessPolicy
		var cfPolicy *cloudflare.AccessPolicy
		var err error
		var action string
		if i < len(current) {
			cfPolicy = &current[i]
		}
		if i < len(expected) {
			k8sPolicy = &expected[i]
		}

		if !cfcollections.AccessPoliciesEqual(cfPolicy, k8sPolicy) {
			if cfPolicy == nil && k8sPolicy != nil {
				action = "create"
				log.Info("accesspolicy is missing - creating...", "policyName", k8sPolicy.Name, "domain", app.Spec.Domain)
				_, err = api.CreateAccessPolicy(ctx, app.Status.AccessApplicationID, *k8sPolicy)
			}
			if k8sPolicy == nil && cfPolicy != nil {
				action = "delete"
				log.Info("accesspolicy is removed - deleting...", "policyId", cfPolicy.ID, "policyName", cfPolicy.Name, "domain", app.Spec.Domain)
				err = api.DeleteAccessPolicy(ctx, app.Status.AccessApplicationID, cfPolicy.ID)
			}
			if cfPolicy != nil && k8sPolicy != nil {
				action = "update"
				k8sPolicy.ID = cfPolicy.ID
				log.Info("accesspolicy is changed - updating...", "policyId", cfPolicy.ID, "policyName", cfPolicy.Name, "domain", app.Spec.Domain)
				_, err = api.UpdateAccessPolicy(ctx, app.Status.AccessApplicationID, *k8sPolicy)
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
		For(&v1alpha1.CloudflareAccessApplication{}, builder.WithPredicates(predicate.GenerationChangedPredicate{})).
		Complete(r)
}
