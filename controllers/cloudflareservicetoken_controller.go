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
	"github.com/bojanzelic/cloudflare-zero-trust-operator/internal/cftypes"
	"github.com/bojanzelic/cloudflare-zero-trust-operator/internal/config"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logger "sigs.k8s.io/controller-runtime/pkg/log"
)

// CloudflareServiceTokenReconciler reconciles a CloudflareServiceToken object.
type CloudflareServiceTokenReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=cloudflare.zelic.io,resources=cloudflareservicetokens,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cloudflare.zelic.io,resources=cloudflareservicetokens/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=cloudflare.zelic.io,resources=cloudflareservicetokens/finalizers,verbs=update

// nolint: gocognit,cyclop
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

	// @todo: change me
	if serviceToken.Status.ServiceTokenID != "" {
		allTokens, err := api.ServiceTokens(ctx)
		if err != nil {
			return ctrl.Result{}, errors.Wrap(err, "unable to create access service token")
		}
		for i, token := range allTokens {
			if token.ID == serviceToken.Status.ServiceTokenID {
				existingServiceToken = &allTokens[i]

				break
			}
		}
	}

	if existingServiceToken == nil {
		token, err := api.CreateAccessServiceToken(ctx, serviceToken.ToExtendedToken())
		existingServiceToken = &token
		if err != nil {
			return ctrl.Result{}, errors.Wrap(err, "unable to create access service token")
		}
	}

	// reconcile  secret
	secret := &corev1.Secret{}

	// @todo: changeme

	// this is used just for populating existingServiceToken
	if serviceToken.Status.SecretRef != nil {
		// we already have a secret created

		// what if it's thre wrong one?
		existingSecretRef := types.NamespacedName{
			Namespace: req.Namespace,
			Name:      serviceToken.Status.SecretRef.Name,
		}

		err = r.Client.Get(ctx, existingSecretRef, secret)
		if err != nil {
			log.Error(err, "Failed to get secret that should exist", "Secret.Name", existingSecretRef.Name)

			return ctrl.Result{}, errors.Wrap(err, "Failed to get Secret")
		}

		// update object with secret ref
		if err := existingServiceToken.SetSecretValues(serviceToken.Status.SecretRef.ClientIDKey, serviceToken.Status.SecretRef.ClientSecretKey, *secret); err != nil {
			return ctrl.Result{}, errors.Wrap(err, "failed to set secret")
		}
	}

	secretNamespacedName := types.NamespacedName{
		Namespace: req.Namespace,
		Name:      req.Name,
	}

	if serviceToken.Spec.Template.Name != "" {
		secretNamespacedName.Name = serviceToken.Spec.Template.Name
	}

	err = r.Client.Get(ctx, secretNamespacedName, secret)
	if err != nil && k8serrors.IsNotFound(err) {
		// create
		secret = &corev1.Secret{}
		secret.Name = secretNamespacedName.Name
		secret.Namespace = secretNamespacedName.Namespace

		secret.Data = map[string][]byte{}
		secret.Data[serviceToken.Spec.Template.ClientSecretKey] = []byte(existingServiceToken.ClientSecret)
		secret.Data[serviceToken.Spec.Template.ClientIDKey] = []byte(existingServiceToken.ClientID)

		err = r.Client.Create(ctx, secret)

		if err != nil {
			log.Error(nil, "failed to create secret", "secret.namespace", secretNamespacedName.Namespace, "secret.name", secretNamespacedName.Name)

			return ctrl.Result{}, errors.Wrap(err, "Failed to create Secret")
		}

		if err := existingServiceToken.SetSecretValues(serviceToken.Spec.Template.ClientIDKey, serviceToken.Spec.Template.ClientSecretKey, *secret); err != nil {
			return ctrl.Result{}, errors.Wrap(err, "failed to set secret")
		}
	} else if err != nil {
		log.Error(err, "failed to retrieve secret", "Secret.Name", secretNamespacedName.Name)

		return ctrl.Result{}, errors.Wrap(err, "Failed to get Secret")
	}

	updatedSecret := secret.DeepCopy()
	updatedSecret.Data[serviceToken.Spec.Template.ClientSecretKey] = []byte(existingServiceToken.ClientSecret)
	updatedSecret.Data[serviceToken.Spec.Template.ClientIDKey] = []byte(existingServiceToken.ClientID)

	if !reflect.DeepEqual(secret, updatedSecret) {
		secret = updatedSecret
		err = r.Client.Update(ctx, secret)
		if err != nil {
			return ctrl.Result{}, errors.Wrap(err, "Failed to update Secret")
		}
	}

	existingServiceToken.SetSecretReference(serviceToken.Spec.Template.ClientIDKey, serviceToken.Spec.Template.ClientSecretKey, *secret)

	if err := ctrl.SetControllerReference(serviceToken, secret, r.Scheme); err != nil {
		return ctrl.Result{}, errors.Wrap(err, "unable to set secret owner reference")
	}

	err = r.ReconcileStatus(ctx, existingServiceToken, serviceToken)
	if err != nil {
		return ctrl.Result{}, errors.Wrap(err, "unable to set status")
	}

	return ctrl.Result{}, nil
}

func (r *CloudflareServiceTokenReconciler) ReconcileStatus(ctx context.Context, cfToken *cftypes.ExtendedServiceToken, k8sToken *v1alpha1.CloudflareServiceToken) error {
	newToken := k8sToken.DeepCopy()

	if cfToken == nil {
		return nil
	}

	newToken.Status.ServiceTokenID = cfToken.ID
	newToken.Status.CreatedAt = v1.NewTime(*cfToken.CreatedAt)
	newToken.Status.UpdatedAt = v1.NewTime(*cfToken.UpdatedAt)
	newToken.Status.ExpiresAt = v1.NewTime(*cfToken.ExpiresAt)
	newToken.Status.SecretRef = &v1alpha1.SecretRef{
		LocalObjectReference: corev1.LocalObjectReference{
			Name: cfToken.K8sSecretRef.SecretName,
		},
		ClientSecretKey: cfToken.K8sSecretRef.ClientSecretKey,
		ClientIDKey:     cfToken.K8sSecretRef.ClientIDKey,
	}

	if !reflect.DeepEqual(newToken.Status, k8sToken.Status) {
		err := r.Status().Update(ctx, newToken)
		if err != nil {
			return errors.Wrap(err, "unable to update token")
		}
	}

	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *CloudflareServiceTokenReconciler) SetupWithManager(mgr ctrl.Manager) error {
	//nolint:wrapcheck
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.CloudflareServiceToken{}).
		Owns(&corev1.Secret{}).
		Complete(r)
}
