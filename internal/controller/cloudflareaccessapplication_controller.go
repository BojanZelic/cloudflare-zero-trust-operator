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
	"reflect"

	v1alpha1 "github.com/bojanzelic/cloudflare-zero-trust-operator/api/v1alpha1"
	"github.com/bojanzelic/cloudflare-zero-trust-operator/internal/cfapi"
	"github.com/bojanzelic/cloudflare-zero-trust-operator/internal/cfcollections"
	"github.com/bojanzelic/cloudflare-zero-trust-operator/internal/cftypes"
	"github.com/bojanzelic/cloudflare-zero-trust-operator/internal/config"
	"github.com/bojanzelic/cloudflare-zero-trust-operator/internal/ctrlhelper"
	"github.com/bojanzelic/cloudflare-zero-trust-operator/internal/services"
	"github.com/pkg/errors"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/event"
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

//nolint:cyclop,gocognit,gocyclo
func (r *CloudflareAccessApplicationReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	var err error
	var existingaccessApp *cftypes.AccessApplication
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

		existingaccessApp = accessApp

		err = r.ReconcileStatus(ctx, accessApp, app)
		if err != nil {
			return ctrl.Result{}, errors.Wrap(err, "issue updating status")
		}
	} else {
		accessApp, err := api.AccessApplication(ctx, app.Status.AccessApplicationID)
		if err != nil {
			if cfapi.IsNotFoundError(err) {
				log.Info("access application not found - recreating...", "accessApplicationID", app.Status.AccessApplicationID)
				app.Status.AccessApplicationID = ""
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
			return ctrl.Result{}, errors.Wrap(err, "unable to create access application")
		}

		if err = r.ReconcileStatus(ctx, &accessapp, app); err != nil {
			return ctrl.Result{}, errors.Wrap(err, "issue updating status")
		}
	}

	if !cfcollections.AccessAppEqual(*existingaccessApp, app.ToCloudflare()) {
		log.Info("app has changed - updating...", "name", app.Spec.Name, "domain", app.Spec.Domain)
		updatedSpec := app.ToCloudflare()
		updatedSpec.PolicyIDs = existingaccessApp.PolicyIDs
		accessapp, err := api.UpdateAccessApplication(ctx, updatedSpec)
		if err != nil {
			return ctrl.Result{}, errors.Wrap(err, "unable to update access application")
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
		return ctrl.Result{}, errors.Wrap(err, "unable to reconcile access policies")
	}

	if _, err = controllerutil.CreateOrPatch(ctx, r.Client, app, func() error {
		meta.SetStatusCondition(&app.Status.Conditions, metav1.Condition{Type: statusAvailable, Status: metav1.ConditionTrue, Reason: "Reconciling", Message: "App Reconciled Successfully"})

		return nil
	}); err != nil {
		return ctrl.Result{}, errors.Wrap(err, "Failed to update CloudflareAccessApplication status")
	}

	return ctrl.Result{}, nil
}

// nolint:dupl
func (r *CloudflareAccessApplicationReconciler) ReconcileStatus(ctx context.Context, cfApp *cftypes.AccessApplication, k8sApp *v1alpha1.CloudflareAccessApplication) error {
	if k8sApp.Status.AccessApplicationID != "" {
		return nil
	}

	if cfApp == nil {
		return nil
	}

	app := k8sApp.DeepCopy()

	if _, err := controllerutil.CreateOrPatch(ctx, r.Client, app, func() error {
		app.Status.AccessApplicationID = cfApp.ID
		if cfApp.CreatedAt != nil {
			app.Status.CreatedAt = metav1.NewTime(*cfApp.CreatedAt)
		}
		if cfApp.UpdatedAt != nil {
			app.Status.UpdatedAt = metav1.NewTime(*cfApp.UpdatedAt)
		}

		return nil
	}); err != nil {
		return errors.Wrap(err, "Failed to update CloudflareAccessApplication status")
	}

	// CreateOrPatch re-fetches the object from k8s which removes any changes we've made that override them
	// so thats why we re-apply these settings again on the original object;
	k8sApp.Status = app.Status

	return nil
}

//nolint:gocognit,cyclop,gocyclo
func (r *CloudflareAccessApplicationReconciler) ReconcilePolicies(ctx context.Context, api *cfapi.API, app *v1alpha1.CloudflareAccessApplication, current, expected cfcollections.AccessPolicyCollection) error {
	log := logger.FromContext(ctx)

	allReusable := true
	for _, p := range current {
		if !p.Reusable {
			allReusable = false

			break
		}
	}

	allEqual := len(current) == len(expected)
	if allEqual {
		for i := range expected {
			if !cfcollections.AccessPoliciesEqual(&current[i], &expected[i]) {
				allEqual = false

				break
			}
		}
	}

	if allReusable && allEqual {
		return nil
	}

	migratedLegacy := false
	for i := range current {
		if !current[i].Reusable {
			migratedLegacy = true
			if i < len(expected) {
				log.Info("migrating legacy application policy to reusable policy", "policyId", current[i].ID, "policyName", current[i].Name, "domain", app.Spec.Domain)
				if err := api.MakeAccessPolicyReusable(ctx, app.Status.AccessApplicationID, current[i].ID); err == nil {
					current[i].Reusable = true
				} else {
					log.Info("make_reusable fallback: replacing legacy application policy", "policyId", current[i].ID, "reason", err.Error())
					if delErr := api.DeleteLegacyApplicationPolicy(ctx, app.Status.AccessApplicationID, current[i].ID); delErr != nil && !cfapi.IsNotFoundError(delErr) {
						return errors.Wrap(delErr, "unable to delete legacy application policy during migration")
					}
					current[i].ID = ""
				}
			} else {
				log.Info("deleting excess legacy application policy", "policyId", current[i].ID, "policyName", current[i].Name, "domain", app.Spec.Domain)
				if err := api.DeleteLegacyApplicationPolicy(ctx, app.Status.AccessApplicationID, current[i].ID); err != nil && !cfapi.IsNotFoundError(err) {
					return errors.Wrap(err, "unable to delete excess legacy application policy")
				}
				current[i].ID = ""
			}
		}
	}

	currentReusableIDs := make([]string, 0, len(current))
	for _, p := range current {
		if p.Reusable && p.ID != "" {
			currentReusableIDs = append(currentReusableIDs, p.ID)
		}
	}

	desiredPolicyIDs := make([]string, 0, len(expected))
	for i := range expected {
		exp := expected[i]
		exp.Precedence = i + 1

		if i < len(current) && current[i].Reusable && current[i].ID != "" {
			if !cfcollections.AccessPolicyContentEqual(&current[i], &exp) {
				polInfo, err := api.GetAccessPolicy(ctx, current[i].ID)
				if err == nil && polInfo.AppCount <= 1 {
					exp.ID = current[i].ID
					log.Info("accesspolicy is changed - updating...", "policyId", exp.ID, "policyName", exp.Name, "domain", app.Spec.Domain)
					if _, err = api.UpdateAccessPolicy(ctx, exp); err != nil {
						return errors.Wrap(err, "Unable to update access policy")
					}
					desiredPolicyIDs = append(desiredPolicyIDs, current[i].ID)

					continue
				}

				if err != nil && !cfapi.IsNotFoundError(err) {
					return errors.Wrap(err, "Unable to inspect access policy")
				}

				log.Info("creating dedicated reusable accesspolicy...", "policyName", exp.Name, "domain", app.Spec.Domain)
				created, createErr := api.CreateAccessPolicy(ctx, exp)
				if createErr != nil {
					return errors.Wrap(createErr, "Unable to create access policy")
				}
				desiredPolicyIDs = append(desiredPolicyIDs, created.ID)

				continue
			}

			desiredPolicyIDs = append(desiredPolicyIDs, current[i].ID)

			continue
		}

		log.Info("accesspolicy is missing - creating...", "policyName", exp.Name, "domain", app.Spec.Domain)
		created, err := api.CreateAccessPolicy(ctx, exp)
		if err != nil {
			return errors.Wrap(err, "Unable to create access policy")
		}
		desiredPolicyIDs = append(desiredPolicyIDs, created.ID)
	}

	precedencesChanged := false
	if len(current) == len(expected) {
		for i := range expected {
			if current[i].Precedence != i+1 {
				precedencesChanged = true

				break
			}
		}
	}

	if migratedLegacy || precedencesChanged || !reflect.DeepEqual(currentReusableIDs, desiredPolicyIDs) {
		if _, err := api.SetAccessApplicationPolicies(ctx, app.ToCloudflare(), desiredPolicyIDs); err != nil {
			return errors.Wrap(err, "Unable to link access policies to application")
		}
	}

	desiredSet := make(map[string]struct{}, len(desiredPolicyIDs))
	for _, id := range desiredPolicyIDs {
		desiredSet[id] = struct{}{}
	}

	for _, curPol := range current {
		if curPol.Reusable && curPol.ID != "" {
			if _, stillUsed := desiredSet[curPol.ID]; !stillUsed {
				polInfo, err := api.GetAccessPolicy(ctx, curPol.ID)
				if err != nil {
					if cfapi.IsNotFoundError(err) {
						continue
					}

					return errors.Wrap(err, "Unable to check removed access policy")
				}

				if polInfo.AppCount == 0 {
					log.Info("accesspolicy is removed - deleting...", "policyId", curPol.ID, "policyName", curPol.Name, "domain", app.Spec.Domain)
					if err := api.DeleteAccessPolicy(ctx, curPol.ID); err != nil && !cfapi.IsNotFoundError(err) {
						return errors.Wrap(err, "Unable to delete access policy")
					}
				}
			}
		}
	}

	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *CloudflareAccessApplicationReconciler) SetupWithManager(mgr ctrl.Manager) error {
	//nolint:wrapcheck
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.CloudflareAccessApplication{}, builder.WithPredicates(predicate.Or(
			predicate.GenerationChangedPredicate{},
			predicate.Funcs{
				UpdateFunc: func(e event.UpdateEvent) bool {
					return e.ObjectOld != nil && e.ObjectNew != nil &&
						e.ObjectOld.GetDeletionTimestamp().IsZero() &&
						!e.ObjectNew.GetDeletionTimestamp().IsZero()
				},
			},
		))).
		Complete(r)
}
