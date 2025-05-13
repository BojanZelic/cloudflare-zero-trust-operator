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

// CloudflareAccessGroupReconciler reconciles a CloudflareAccessGroup object.
type CloudflareAccessGroupReconciler struct {
	client.Client
	Scheme         *runtime.Scheme
	Helper         *ctrlhelper.ControllerHelper
	OptionalTracer *cfapi.InsertedCFRessourcesTracer
}

// +kubebuilder:rbac:groups=cloudflare.zelic.io,resources=cloudflareaccessgroups,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cloudflare.zelic.io,resources=cloudflareaccessgroups/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=cloudflare.zelic.io,resources=cloudflareaccessgroups/finalizers,verbs=update

//nolint:cyclop,gocognit
func (r *CloudflareAccessGroupReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	var err error
	var existingCfAG *zero_trust.AccessGroupGetResponse
	var api *cfapi.API

	log := logger.FromContext(ctx).WithName("CloudflareAccessGroupController")

	accessGroup := &v4alpha1.CloudflareAccessGroup{}

	err = r.Get(ctx, req.NamespacedName, accessGroup)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}

		log.Error(err, "Failed to get CloudflareAccessGroup", "CloudflareAccessGroup.Name", req.Name)

		return ctrl.Result{}, errors.Wrap(err, "Failed to get CloudflareAccessGroup")
	}

	cfConfig := config.ParseCloudflareConfig(accessGroup)
	validConfig, err := cfConfig.IsValid()
	if !validConfig {
		return ctrl.Result{}, errors.Wrap(err, "invalid config")
	}

	api = cfapi.New(cfConfig.APIToken, cfConfig.APIKey, cfConfig.APIEmail, cfConfig.AccountID, r.OptionalTracer)

	continueReconcilliation, err := r.Helper.ReconcileDeletion(ctx, api, accessGroup)
	if !continueReconcilliation || err != nil {
		if err != nil {
			log.Error(err, "unable to reconcile deletion for access group")
		}

		return ctrl.Result{}, errors.Wrap(err, "unable to reconcile deletion")
	}

	_, err = controllerutil.CreateOrPatch(ctx, r.Client, accessGroup, func() error {
		if len(accessGroup.Status.Conditions) == 0 {
			meta.SetStatusCondition(&accessGroup.Status.Conditions,
				metav1.Condition{
					Type:    statusAvailable,
					Status:  metav1.ConditionUnknown,
					Reason:  "Reconciling",
					Message: "AccessGroup is reconciling",
				},
			)
		}

		return nil
	})

	if err != nil {
		return ctrl.Result{}, errors.Wrap(err, "Failed to update CloudflareAccessGroup status")
	}

	if accessGroup.Status.AccessGroupID == "" {
		existingCfAG, err = api.AccessGroupByName(ctx, accessGroup.Spec.Name)
		if err != nil {
			return ctrl.Result{}, errors.Wrap(err, "unable to get access group")
		}
		if existingCfAG != nil {
			log.Info("access group already exists. importing...", "accessGroup", existingCfAG.Name, "accessGroupID", existingCfAG.ID)
		}
		err = r.ReconcileStatus(ctx, existingCfAG, accessGroup)
		if err != nil {
			return ctrl.Result{}, errors.Wrap(err, "unable to update access groups")
		}
	} else {
		existingCfAG, err = api.AccessGroup(ctx, accessGroup.Status.AccessGroupID)
		if err != nil {
			return ctrl.Result{}, errors.Wrap(err, "unable to get access groups")
		}
	}

	//
	if popRes, err := v4alpha1.PopulateWithCloudflareUUIDs(ctx, req.Namespace, r.Client, accessGroup); err != nil {
		_, err = controllerutil.CreateOrPatch(ctx, r.Client, accessGroup, func() error {
			meta.SetStatusCondition(&accessGroup.Status.Conditions,
				metav1.Condition{
					Type:    statusDegraded,
					Status:  metav1.ConditionFalse,
					Reason:  "InvalidReference",
					Message: err.Error(),
				},
			)

			return nil
		})

		//
		log.Info("failed to update access group's referenced CloudFlare UUIDs")
		if err != nil {
			return ctrl.Result{}, errors.Wrap(err, "Failed to update CloudflareAccessGroup status")
		}

		// might requeue !
		return popRes, nil
	}

	if existingCfAG == nil {
		//nolint:varnamelen
		existingCfAG, err = api.CreateAccessGroup(ctx, accessGroup)
		if err != nil {
			return ctrl.Result{}, errors.Wrap(err, "unable to create access group")
		}
		err = r.ReconcileStatus(ctx, existingCfAG, accessGroup)
		if err != nil {
			return ctrl.Result{}, errors.Wrap(err, "unable to set access group status")
		}
	} else if needsUpdate := !cfcompare.AreAccessGroupsEquivalent(existingCfAG, accessGroup); needsUpdate {
		log.Info(accessGroup.Spec.Name + " has changed, updating...")

		err := api.UpdateAccessGroup(ctx, accessGroup)
		if err != nil {
			return ctrl.Result{}, errors.Wrap(err, "unable to update access groups")
		}
	}

	_, err = controllerutil.CreateOrPatch(ctx, r.Client, accessGroup, func() error {
		meta.SetStatusCondition(&accessGroup.Status.Conditions,
			metav1.Condition{
				Type:    statusAvailable,
				Status:  metav1.ConditionTrue,
				Reason:  "Reconcilied",
				Message: "AccessGroup Reconciled Successfully",
			},
		)

		return nil
	})

	if err != nil {
		return ctrl.Result{}, errors.Wrap(err, "Failed to update CloudflareAccessGroup status")
	}

	log.Info("reconciled successfully")

	return ctrl.Result{}, nil
}

//nolint:dupl
func (r *CloudflareAccessGroupReconciler) ReconcileStatus(ctx context.Context, cfGroup *zero_trust.AccessGroupGetResponse, k8sGroup *v4alpha1.CloudflareAccessGroup) error {
	if k8sGroup.Status.AccessGroupID != "" {
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
		return errors.Wrap(err, "Failed to update CloudflareAccessGroup status")
	}

	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *CloudflareAccessGroupReconciler) SetupWithManager(mgr ctrl.Manager) error {
	//nolint:wrapcheck
	return ctrl.NewControllerManagedBy(mgr).
		For(&v4alpha1.CloudflareAccessGroup{}, builder.WithPredicates(predicate.GenerationChangedPredicate{})).
		Complete(r)
}
