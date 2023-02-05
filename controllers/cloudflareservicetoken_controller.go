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
	v1alpha1 "github.com/bojanzelic/cloudflare-zero-trust-operator/api/v1alpha1/v1alpha1"
	"github.com/bojanzelic/cloudflare-zero-trust-operator/internal/cfapi"
	"github.com/bojanzelic/cloudflare-zero-trust-operator/internal/cftypes"
	"github.com/bojanzelic/cloudflare-zero-trust-operator/internal/config"
	"github.com/bojanzelic/cloudflare-zero-trust-operator/internal/ctrlhelper"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	Helper *ctrlhelper.ControllerHelper
}

// +kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cloudflare.zelic.io,resources=cloudflareservicetokens,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cloudflare.zelic.io,resources=cloudflareservicetokens/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=cloudflare.zelic.io,resources=cloudflareservicetokens/finalizers,verbs=update

// nolint: gocognit,cyclop,gocyclo,maintidx
func (r *CloudflareServiceTokenReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	var err error
	var existingServiceToken *cftypes.ExtendedServiceToken
	var api *cfapi.API

	log := logger.FromContext(ctx).WithName("CloudflareServiceTokenController").WithValues(
		"type", "CloudflareAccessApplication",
		"name", req.Name,
		"namespace", req.Namespace,
	)

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

	continueReconcilliation, err := r.Helper.ReconcileDeletion(ctx, api, serviceToken)
	if !continueReconcilliation || err != nil {
		if err != nil {
			log.Error(err, "unable to reconcile deletion for service token")
		}

		return ctrl.Result{}, errors.Wrap(err, "unable to reconcile deletion")
	}

	if serviceToken.Status.Conditions == nil || len(serviceToken.Status.Conditions) == 0 {
		meta.SetStatusCondition(&serviceToken.Status.Conditions, metav1.Condition{Type: statusAvailable, Status: metav1.ConditionUnknown, Reason: "Reconciling", Message: "ServiceToken is reconciling"})
		if err = r.Status().Update(ctx, serviceToken); err != nil {
			return ctrl.Result{}, errors.Wrap(err, "Failed to update CloudflareServiceToken status")
		}

		// refetch the serviceToken
		if err = r.Client.Get(ctx, req.NamespacedName, serviceToken); err != nil {
			return ctrl.Result{}, errors.Wrap(err, "Failed to re-fetch CloudflareServiceToken")
		}
	}

	// this is used just for populating existingServiceToken
	secretList := &corev1.SecretList{}
	if err := r.Client.List(ctx, secretList,
		client.MatchingLabels{v1alpha1.LabelOwnedBy: serviceToken.Name},
		client.InNamespace(serviceToken.Namespace),
	); err != nil {
		return ctrl.Result{}, errors.Wrap(err, "unable to list created secrets")
	}

	secret := &corev1.Secret{}

	if len(secretList.Items) > 0 {
		// we already have a secret created
		secret = &secretList.Items[0]

		if len(secretList.Items) > 1 {
			log.Info("Found multiple secrets with the same owner label", "label", v1alpha1.LabelOwnedBy, "owner", serviceToken.Name)
		}
	}

	if !secret.CreationTimestamp.IsZero() {
		allTokens, err := api.ServiceTokens(ctx)
		if err != nil {
			return ctrl.Result{}, errors.Wrap(err, "unable to create access service token")
		}
		for i, token := range allTokens {
			if token.ID == string(secret.Data[secret.Annotations[v1alpha1.AnnotationTokenIDKey]]) {
				existingServiceToken = &allTokens[i]

				break
			}
		}
	}

	if existingServiceToken == nil {
		token, err := api.CreateAccessServiceToken(ctx, serviceToken.ToExtendedToken())
		log.Info("created access service token", "token_id", token.ID)
		existingServiceToken = &token
		if err != nil {
			return ctrl.Result{}, errors.Wrap(err, "unable to create access service token")
		}
	}

	// update object with secret ref
	if !secret.CreationTimestamp.IsZero() {
		if err := existingServiceToken.SetSecretValues(*secret); err != nil {
			return ctrl.Result{}, errors.Wrap(err, "failed to set secret")
		}
	}

	// reconcile  secret
	secretNamespacedName := types.NamespacedName{
		Namespace: req.Namespace,
		Name:      req.Name,
	}

	if serviceToken.Spec.Template.Name != "" {
		secretNamespacedName.Name = serviceToken.Spec.Template.Name
	}

	var secretToDelete *corev1.Secret
	// secret exists & was renamed; remove the old one
	if !secret.CreationTimestamp.IsZero() && secretNamespacedName.Name != secret.Name {
		secretToDelete = secret
	}

	err = r.Client.Get(ctx, secretNamespacedName, secret)
	//nolint:nestif
	if err != nil && k8serrors.IsNotFound(err) {
		// create
		secret = &corev1.Secret{}
		secret.Name = secretNamespacedName.Name
		secret.Namespace = secretNamespacedName.Namespace
		secret.SetLabels(map[string]string{
			v1alpha1.LabelOwnedBy: serviceToken.Name,
		})
		secret.SetAnnotations(map[string]string{
			v1alpha1.AnnotationClientIDKey:     serviceToken.Spec.Template.ClientIDKey,
			v1alpha1.AnnotationClientSecretKey: serviceToken.Spec.Template.ClientSecretKey,
			v1alpha1.AnnotationTokenIDKey:      "serviceTokenID",
		})

		secret.Data = map[string][]byte{}
		secret.Data[serviceToken.Spec.Template.ClientSecretKey] = []byte(existingServiceToken.ClientSecret)
		secret.Data[serviceToken.Spec.Template.ClientIDKey] = []byte(existingServiceToken.ClientID)
		secret.Data["serviceTokenID"] = []byte(existingServiceToken.ID)

		if err = r.Client.Create(ctx, secret); err != nil {
			log.Error(nil, "failed to create secret", "secret.namespace", secretNamespacedName.Namespace, "secret.name", secretNamespacedName.Name)

			return ctrl.Result{}, errors.Wrap(err, "Failed to create Secret")
		}

		log.Info("created secret", "secret.namespace", secretNamespacedName.Namespace, "secret.name", secretNamespacedName.Name)

		if secretToDelete != nil {
			if err := r.Client.Delete(ctx, secretToDelete); err != nil {
				log.Error(nil, "failed to remove secret", "secret.namespace", secretToDelete.Namespace, "secret.name", secretToDelete.Name)
			}

			log.Info("removed secret", "secret.namespace", secretNamespacedName.Namespace, "secret.name", secretNamespacedName.Name)
		}

		if err := existingServiceToken.SetSecretValues(*secret); err != nil {
			return ctrl.Result{}, errors.Wrap(err, "failed to set secret")
		}
	} else if err != nil {
		log.Error(err, "failed to retrieve secret", "Secret.Name", secretNamespacedName.Name)

		return ctrl.Result{}, errors.Wrap(err, "Failed to get Secret")
	}

	updatedSecret := secret.DeepCopy()
	updatedSecret.SetAnnotations(map[string]string{
		v1alpha1.AnnotationClientIDKey:     serviceToken.Spec.Template.ClientIDKey,
		v1alpha1.AnnotationClientSecretKey: serviceToken.Spec.Template.ClientSecretKey,
		v1alpha1.AnnotationTokenIDKey:      "serviceTokenID",
	})

	updatedSecret.Data[serviceToken.Spec.Template.ClientSecretKey] = []byte(existingServiceToken.ClientSecret)
	updatedSecret.Data[serviceToken.Spec.Template.ClientIDKey] = []byte(existingServiceToken.ClientID)
	updatedSecret.Data["serviceTokenID"] = []byte(existingServiceToken.ID)
	if !reflect.DeepEqual(secret, updatedSecret) {
		secret = updatedSecret
		err = r.Client.Update(ctx, secret)
		if err != nil {
			return ctrl.Result{}, errors.Wrap(err, "Failed to update Secret")
		}

		log.Info("updated secret", "secret.namespace", secretNamespacedName.Namespace, "secret.name", secretNamespacedName.Name)
	}

	existingServiceToken.SetSecretReference(serviceToken.Spec.Template.ClientIDKey, serviceToken.Spec.Template.ClientSecretKey, *secret)

	if err := ctrl.SetControllerReference(serviceToken, secret, r.Scheme); err != nil {
		return ctrl.Result{}, errors.Wrap(err, "unable to set secret owner reference")
	}

	err = r.ReconcileStatus(ctx, existingServiceToken, serviceToken)
	if err != nil {
		return ctrl.Result{}, errors.Wrap(err, "unable to set status")
	}

	if err = r.Client.Get(ctx, req.NamespacedName, serviceToken); err != nil {
		return ctrl.Result{}, errors.Wrap(err, "Failed to re-fetch CloudflareServiceToken")
	}

	meta.SetStatusCondition(&serviceToken.Status.Conditions, metav1.Condition{Type: statusAvailable, Status: metav1.ConditionTrue, Reason: "Reconciling", Message: "CloudflareServiceToken Reconciled Successfully"})
	if err = r.Status().Update(ctx, serviceToken); err != nil {
		return ctrl.Result{}, errors.Wrap(err, "Failed to update CloudflareServiceToken status")
	}

	return ctrl.Result{}, nil
}

func (r *CloudflareServiceTokenReconciler) ReconcileStatus(ctx context.Context, cfToken *cftypes.ExtendedServiceToken, k8sToken *v1alpha1.CloudflareServiceToken) error {
	newToken := k8sToken.DeepCopy()

	if cfToken == nil {
		return nil
	}

	newToken.Status.ServiceTokenID = cfToken.ID
	newToken.Status.CreatedAt = metav1.NewTime(*cfToken.CreatedAt)
	newToken.Status.UpdatedAt = metav1.NewTime(*cfToken.UpdatedAt)
	newToken.Status.ExpiresAt = metav1.NewTime(*cfToken.ExpiresAt)
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

		// refetch the serviceToken
		if err = r.Client.Get(ctx, types.NamespacedName{Namespace: k8sToken.Namespace, Name: k8sToken.Name}, k8sToken); err != nil {
			return errors.Wrap(err, "Failed to re-fetch CloudflareServiceToken")
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
