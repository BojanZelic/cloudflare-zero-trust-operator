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

	v1alpha1 "github.com/bojanzelic/cloudflare-zero-trust-operator/api/v1alpha1"
	"github.com/bojanzelic/cloudflare-zero-trust-operator/internal/cfapi"
	"github.com/bojanzelic/cloudflare-zero-trust-operator/internal/config"
	cloudflare "github.com/cloudflare/cloudflare-go"
	"github.com/pkg/errors"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
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

func (r *CloudflareAccessGroupReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	var err error
	var existingCfAG *cloudflare.AccessGroup
	var api *cfapi.API

	log := logger.FromContext(ctx)
	ag := &v1alpha1.CloudflareAccessGroup{}

	err = r.Client.Get(ctx, req.NamespacedName, ag)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}

		log.Error(err, "Failed to get CloudflareAccessGroup", "CloudflareAccessGroup.Name", ag.Name)

		return ctrl.Result{}, errors.Wrap(err, "Failed to get CloudflareAccessGroup")
	}

	cfConfig := config.ParseCloudflareConfig(ag)
	validConfig, err := cfConfig.IsValid()
	if !validConfig {
		return ctrl.Result{}, errors.Wrap(err, "invalid config")
	}

	api, err = cfapi.New(cfConfig.APIToken, cfConfig.APIKey, cfConfig.APIEmail, cfConfig.AccountID)

	if err != nil {
		return ctrl.Result{}, errors.Wrap(err, "unable to initialize cloudflare object")
	}

	// Fetch user details on the account
	// @todo paginate
	cfAccessGroups, err := api.AccessGroups(ctx)
	if err != nil {
		return ctrl.Result{}, errors.Wrap(err, "unable to get access groups")
	}

	newCfAG := ag.ToCloudflare()

	if ag.Status.AccessGroupID == "" {
		existingCfAG = cfAccessGroups.GetByName(ag.CloudflareName())

		if existingCfAG != nil {
			log.Info(ag.CloudflareName() + " already exists")

			ag.Status.AccessGroupID = existingCfAG.ID
			ag.Status.CreatedAt = v1.NewTime(*existingCfAG.CreatedAt)
			ag.Status.UpdatedAt = v1.NewTime(*existingCfAG.UpdatedAt)

			newCfAG = ag.ToCloudflare()

			err := r.Status().Update(ctx, ag)
			if err != nil {
				return ctrl.Result{}, errors.Wrap(err, "unable to update access group")
			}
		}
	} else {
		cfAG, err := api.AccessGroup(ctx, ag.Status.AccessGroupID)
		existingCfAG = &cfAG
		if err != nil {
			return ctrl.Result{}, errors.Wrap(err, "unable to get access groups")
		}
	}

	if existingCfAG == nil {
		_, err = api.CreateAccessGroup(ctx, newCfAG)
		if err != nil {
			return ctrl.Result{}, errors.Wrap(err, "unable to create access group")
		}
	}

	if !cfapi.AcessGroupEmailEqual(*existingCfAG, newCfAG) {
		log.Info(newCfAG.Name + " has changed, updating...")

		_, err := api.UpdateAccessGroup(ctx, newCfAG)
		if err != nil {
			return ctrl.Result{}, errors.Wrap(err, "unable to update access groups")
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *CloudflareAccessGroupReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.CloudflareAccessGroup{}).
		Complete(r)
}
