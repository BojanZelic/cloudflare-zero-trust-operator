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

	v4alpha1 "github.com/bojanzelic/cloudflare-zero-trust-operator/api/v4alpha1"
	"github.com/bojanzelic/cloudflare-zero-trust-operator/internal/cfapi"
	"github.com/bojanzelic/cloudflare-zero-trust-operator/internal/cfcompare"
	"github.com/bojanzelic/cloudflare-zero-trust-operator/internal/config"
	"github.com/bojanzelic/cloudflare-zero-trust-operator/internal/ctrlhelper"
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

// CloudflareAccessReusablePolicyReconciler reconciles a CloudflareAccessReusablePolicy object.
type CloudflareAccessReusablePolicyReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Helper *ctrlhelper.ControllerHelper

	// Mainly used for debug / tests purposes. Should not be instantiated in production run.
	OptionalTracer *cfapi.InsertedCFRessourcesTracer
}

// +kubebuilder:rbac:groups=cloudflare.zelic.io,resources=cloudflareaccessreusablepolicies,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cloudflare.zelic.io,resources=cloudflareaccessreusablepolicies/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=cloudflare.zelic.io,resources=cloudflareaccessreusablepolicies/finalizers,verbs=update

//nolint:cyclop,gocognit
func (r *CloudflareAccessReusablePolicyReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	var err error
	var existingCfRP *zero_trust.AccessPolicyGetResponse
	var api *cfapi.API

	log := logger.FromContext(ctx).WithName("CloudflareAccessReusablePolicyController")

	reusablePolicy := &v4alpha1.CloudflareAccessReusablePolicy{}

	err = r.Get(ctx, req.NamespacedName, reusablePolicy)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			// will stop
			return ctrl.Result{}, nil
		}

		log.Error(err, "Failed to get CloudflareAccessReusablePolicy", "CloudflareAccessReusablePolicy.Name", req.Name)

		// will retry immediately
		return ctrl.Result{}, errors.Wrap(err, "Failed to get CloudflareAccessReusablePolicy")
	}

	cfConfig := config.ParseCloudflareConfig(reusablePolicy)
	validConfig, err := cfConfig.IsValid()
	if !validConfig {
		// will retry immediately
		return ctrl.Result{}, errors.Wrap(err, "invalid config")
	}

	api = cfapi.New(cfConfig.APIToken, cfConfig.APIKey, cfConfig.APIEmail, cfConfig.AccountID, r.OptionalTracer)

	continueReconcilliation, err := r.Helper.ReconcileDeletion(ctx, api, reusablePolicy)
	if !continueReconcilliation || err != nil {
		if err != nil {
			log.Error(err, "unable to reconcile deletion for reusable policy")
		}

		// will retry immediately
		return ctrl.Result{}, errors.Wrap(err, "unable to reconcile deletion")
	}

	_, err = controllerutil.CreateOrPatch(ctx, r.Client, reusablePolicy, func() error {
		if len(reusablePolicy.Status.Conditions) == 0 {
			meta.SetStatusCondition(&reusablePolicy.Status.Conditions,
				metav1.Condition{
					Type:    statusAvailable,
					Status:  metav1.ConditionUnknown,
					Reason:  "Reconciling",
					Message: "AccessReusablePolicy is reconciling",
				},
			)
		}

		return nil
	})

	if err != nil {
		// will retry immediately
		return ctrl.Result{}, errors.Wrap(err, "Failed to update CloudflareAccessReusablePolicy status")
	}

	if reusablePolicy.Status.AccessReusablePolicyID != "" {
		existingCfRP, err = api.AccessReusablePolicy(ctx, reusablePolicy.Status.AccessReusablePolicyID)
		if err != nil {
			// will retry immediately
			return ctrl.Result{}, errors.Wrap(err, "unable to get reusable policies")
		}
	}

	//
	//
	//

	// populate UUIDs; if failure:
	popRes, err := v4alpha1.PopulateWithCloudflareUUIDs(ctx, req.Namespace, &log, r.Client, reusablePolicy)

	// if any result returned, return it to reconcilier along w/ err (if any)
	if popRes != nil {
		return *popRes, errors.Wrap(err, "Failed to populate CF UUIDs")
	} else if err != nil {
		//
		log.Info("failed to update access policy's referenced CloudFlare UUIDs")

		// patch with status updated
		_, errPatch := controllerutil.CreateOrPatch(ctx, r.Client, reusablePolicy, func() error {
			meta.SetStatusCondition(&reusablePolicy.Status.Conditions,
				metav1.Condition{
					Type:    statusDegraded,
					Status:  metav1.ConditionFalse,
					Reason:  "InvalidReference",
					Message: err.Error(),
				},
			)
			return nil
		})

		if errPatch != nil {
			// will retry immediately
			return ctrl.Result{}, errors.Wrap(errPatch, "Failed to update CloudflareAccessReusablePolicy Status, after a CF UUIDs population failure")
		}

		// will retry immediately
		return ctrl.Result{}, errors.Wrap(err, "Failed to populate CF UUIDs")
	}

	//
	//
	//

	if existingCfRP == nil {
		//nolint:varnamelen
		existingCfRP, err = api.CreateAccessReusablePolicy(ctx, reusablePolicy)
		if err != nil {
			// will retry immediately
			return ctrl.Result{}, errors.Wrap(err, "unable to create reusable policy")
		}
		err = r.ReconcileStatus(ctx, existingCfRP, reusablePolicy)
		if err != nil {
			// will retry immediately
			return ctrl.Result{}, errors.Wrap(err, "unable to set reusable policy status")
		}
	} else if mustUpdate := !cfcompare.AreAccessReusablePoliciesEquivalent(existingCfRP, reusablePolicy); mustUpdate {
		log.Info(reusablePolicy.Name + " diverge from remote counterpart, updating CF API...")

		err := api.UpdateAccessReusablePolicy(ctx, reusablePolicy)
		if err != nil {
			// will retry immediately
			return ctrl.Result{}, errors.Wrap(err, "unable to update reusable policies")
		}
	}

	_, err = controllerutil.CreateOrPatch(ctx, r.Client, reusablePolicy, func() error {
		meta.SetStatusCondition(&reusablePolicy.Status.Conditions,
			metav1.Condition{
				Type:    statusAvailable,
				Status:  metav1.ConditionTrue,
				Reason:  "Reconcilied",
				Message: "AccessPolicy Reconciled Successfully",
			},
		)

		return nil
	})

	if err != nil {
		// will retry immediately
		return ctrl.Result{}, errors.Wrap(err, "Failed to update CloudflareAccessReusablePolicy status")
	}

	log.Info("reconciled successfully")

	// will stop normally
	return ctrl.Result{}, nil
}

//nolint:dupl
func (r *CloudflareAccessReusablePolicyReconciler) ReconcileStatus(
	ctx context.Context,
	cfReusablPolicy *zero_trust.AccessPolicyGetResponse,
	k8sReusablePolicy *v4alpha1.CloudflareAccessReusablePolicy,
) error {
	if k8sReusablePolicy.Status.AccessReusablePolicyID != "" {
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
		return errors.Wrap(err, "Failed to update CloudflareAccessReusablePolicy status")
	}

	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *CloudflareAccessReusablePolicyReconciler) SetupWithManager(mgr ctrl.Manager) error {
	//nolint:wrapcheck
	return ctrl.NewControllerManagedBy(mgr).
		For(&v4alpha1.CloudflareAccessReusablePolicy{}, builder.WithPredicates(predicate.GenerationChangedPredicate{})).
		Complete(r)
}
