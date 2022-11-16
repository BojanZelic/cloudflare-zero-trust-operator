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
	"fmt"

	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logger "sigs.k8s.io/controller-runtime/pkg/log"

	v1alpha1 "github.com/bojanzelic/cloudflare-zero-trust-operator/api/v1alpha1"
	"github.com/bojanzelic/cloudflare-zero-trust-operator/internal/cfapi"
	"github.com/bojanzelic/cloudflare-zero-trust-operator/internal/cftypes"
	"github.com/bojanzelic/cloudflare-zero-trust-operator/internal/config"
	"github.com/pkg/errors"
)

// CloudflareServiceTokenReconciler reconciles a CloudflareServiceToken object.
type CloudflareServiceTokenReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=cloudflare.zelic.io,resources=cloudflareservicetokens,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=cloudflare.zelic.io,resources=cloudflareservicetokens/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=cloudflare.zelic.io,resources=cloudflareservicetokens/finalizers,verbs=update

func (r *CloudflareServiceTokenReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	var err error
	var existingServiceToken *cftypes.ExtendedServiceToken
	var api *cfapi.API

	log := logger.FromContext(ctx)
	serviceToken := &v1alpha1.CloudflareServiceToken{}

	err = r.Client.Get(ctx, req.NamespacedName, serviceToken)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}

		log.Error(err, "Failed to get CloudflareServiceToken", "CloudflareServiceToken.Name", req.Name)

		return ctrl.Result{}, errors.Wrap(err, "Failed to get CloudflareServiceToken")
	}

	cfConfig := config.ParseCloudflareConfig(serviceToken)
	validConfig, err := cfConfig.IsValid()
	if !validConfig {
		return ctrl.Result{}, errors.Wrap(err, "invalid config")
	}

	api, err = cfapi.New(cfConfig.APIToken, cfConfig.APIKey, cfConfig.APIEmail, cfConfig.AccountID)

	if err != nil {
		return ctrl.Result{}, errors.Wrap(err, "unable to initialize cloudflare object")
	}

	if serviceToken.Status.ServiceTokenID == "" {
		// create
		token, err := api.CreateAccessServiceToken(ctx, serviceToken.ToExtendedToken())
		existingServiceToken = &token

		fmt.Println(existingServiceToken)
		if err != nil {
			return ctrl.Result{}, errors.Wrap(err, "unable to create access service token")
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *CloudflareServiceTokenReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.CloudflareServiceToken{}).
		Owns(&corev1.Secret{}).
		Complete(r)
}
