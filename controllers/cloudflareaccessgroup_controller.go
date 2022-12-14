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

package controllers

import (
	"context"
	"reflect"

	v1alpha1 "github.com/bojanzelic/cloudflare-zero-trust-operator/api/v1alpha1"
	"github.com/bojanzelic/cloudflare-zero-trust-operator/internal/cfapi"
	"github.com/bojanzelic/cloudflare-zero-trust-operator/internal/cfcollections"
	"github.com/bojanzelic/cloudflare-zero-trust-operator/internal/config"
	cloudflare "github.com/cloudflare/cloudflare-go"
	"github.com/pkg/errors"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logger "sigs.k8s.io/controller-runtime/pkg/log"
)

// CloudflareAccessGroupReconciler reconciles a CloudflareAccessGroup object.
type CloudflareAccessGroupReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=cloudflare.zelic.io,resources=cloudflareaccessgroups,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=cloudflare.zelic.io,resources=cloudflareaccessgroups/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=cloudflare.zelic.io,resources=cloudflareaccessgroups/finalizers,verbs=update

//nolint:cyclop,gocognit
func (r *CloudflareAccessGroupReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	var err error
	var existingCfAG *cloudflare.AccessGroup
	var api *cfapi.API

	log := logger.FromContext(ctx).WithName("CloudflareAccessGroupController")
	accessGroup := &v1alpha1.CloudflareAccessGroup{}

	err = r.Client.Get(ctx, req.NamespacedName, accessGroup)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}

		log.Error(err, "Failed to get CloudflareAccessGroup", "CloudflareAccessGroup.Name", req.Name)

		return ctrl.Result{}, errors.Wrap(err, "Failed to get CloudflareAccessGroup")
	}

	if accessGroup.Status.Conditions == nil || len(accessGroup.Status.Conditions) == 0 {
		meta.SetStatusCondition(&accessGroup.Status.Conditions, metav1.Condition{Type: statusAvailable, Status: metav1.ConditionUnknown, Reason: "Reconciling", Message: "AccessGroup is reconciling"})
		if err = r.Status().Update(ctx, accessGroup); err != nil {
			return ctrl.Result{}, errors.Wrap(err, "Failed to update AccessGroup status")
		}

		// refetch the group
		if err = r.Client.Get(ctx, req.NamespacedName, accessGroup); err != nil {
			return ctrl.Result{}, errors.Wrap(err, "Failed to re-fetch CloudflareAccessGroup")
		}
	}

	cfConfig := config.ParseCloudflareConfig(accessGroup)
	validConfig, err := cfConfig.IsValid()
	if !validConfig {
		return ctrl.Result{}, errors.Wrap(err, "invalid config")
	}

	api, err = cfapi.New(cfConfig.APIToken, cfConfig.APIKey, cfConfig.APIEmail, cfConfig.AccountID)

	if err != nil {
		return ctrl.Result{}, errors.Wrap(err, "unable to initialize cloudflare object")
	}

	cfAccessGroups, err := api.AccessGroups(ctx)
	if err != nil {
		return ctrl.Result{}, errors.Wrap(err, "unable to get access groups")
	}

	newCfAG := accessGroup.ToCloudflare()

	if accessGroup.Status.AccessGroupID == "" {
		existingCfAG = cfAccessGroups.GetByName(accessGroup.Spec.Name)
		if existingCfAG != nil {
			log.Info("access group already exists. importing...", "accessGroup", existingCfAG.Name, "accessGroupID", existingCfAG.ID)
		}
		err = r.ReconcileStatus(ctx, existingCfAG, accessGroup)
		if err != nil {
			return ctrl.Result{}, errors.Wrap(err, "unable to update access groups")
		}
	} else {
		cfAG, err := api.AccessGroup(ctx, accessGroup.Status.AccessGroupID)
		existingCfAG = &cfAG
		if err != nil {
			return ctrl.Result{}, errors.Wrap(err, "unable to get access groups")
		}
	}

	if existingCfAG == nil {
		//nolint:varnamelen
		ag, err := api.CreateAccessGroup(ctx, newCfAG)
		if err != nil {
			return ctrl.Result{}, errors.Wrap(err, "unable to create access group")
		}
		err = r.ReconcileStatus(ctx, &ag, accessGroup)
		if err != nil {
			return ctrl.Result{}, errors.Wrap(err, "unable to set access group status")
		}
		existingCfAG = &ag
	}

	if !cfcollections.AccessGroupEqual(*existingCfAG, newCfAG) {
		log.Info(newCfAG.Name + " has changed, updating...")

		_, err := api.UpdateAccessGroup(ctx, newCfAG)
		if err != nil {
			return ctrl.Result{}, errors.Wrap(err, "unable to update access groups")
		}
	}

	err = r.Client.Get(ctx, req.NamespacedName, accessGroup)
	if err != nil {
		return ctrl.Result{}, errors.Wrap(err, "Failed to re-fetch CloudflareAccessGroup")
	}

	meta.SetStatusCondition(&accessGroup.Status.Conditions, metav1.Condition{Type: statusAvailable, Status: metav1.ConditionTrue, Reason: "Reconciling", Message: "AccessGroup Reconciled Successfully"})
	if err = r.Status().Update(ctx, accessGroup); err != nil {
		return ctrl.Result{}, errors.Wrap(err, "Failed to update CloudflareAccessGroup status")
	}

	return ctrl.Result{}, nil
}

// nolint:dupl
func (r *CloudflareAccessGroupReconciler) ReconcileStatus(ctx context.Context, cfGroup *cloudflare.AccessGroup, k8sGroup *v1alpha1.CloudflareAccessGroup) error {
	if k8sGroup.Status.AccessGroupID != "" {
		return nil
	}

	if cfGroup == nil {
		return nil
	}

	newGroup := k8sGroup.DeepCopy()

	newGroup.Status.AccessGroupID = cfGroup.ID
	newGroup.Status.CreatedAt = metav1.NewTime(*cfGroup.CreatedAt)
	newGroup.Status.UpdatedAt = metav1.NewTime(*cfGroup.UpdatedAt)

	if !reflect.DeepEqual(k8sGroup.Status, newGroup.Status) {
		err := r.Status().Update(ctx, newGroup)
		if err != nil {
			return errors.Wrap(err, "unable to update access group")
		}

		namespacedName := types.NamespacedName{Name: k8sGroup.Name, Namespace: k8sGroup.Namespace}
		// refetch the group
		if err = r.Client.Get(ctx, namespacedName, k8sGroup); err != nil {
			return errors.Wrap(err, "Failed to re-fetch CloudflareAccessGroup")
		}
	}

	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *CloudflareAccessGroupReconciler) SetupWithManager(mgr ctrl.Manager) error {
	//nolint:wrapcheck
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.CloudflareAccessGroup{}).
		Complete(r)
}
