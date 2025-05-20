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

	"github.com/Southclaws/fault"
	"github.com/Southclaws/fault/fctx"
	"github.com/Southclaws/fault/fmsg"
	v4alpha1 "github.com/bojanzelic/cloudflare-zero-trust-operator/api/v4alpha1"
	"github.com/bojanzelic/cloudflare-zero-trust-operator/internal/cfapi"
	"github.com/bojanzelic/cloudflare-zero-trust-operator/internal/cfcompare"
	"github.com/bojanzelic/cloudflare-zero-trust-operator/internal/config"
	"github.com/bojanzelic/cloudflare-zero-trust-operator/internal/ctrlhelper"
	"github.com/cloudflare/cloudflare-go/v4/zero_trust"
	"github.com/go-logr/logr"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	SearchOnAppTypeIndex string = "spec.type"
)

// CloudflareAccessApplicationReconciler reconciles a CloudflareAccessApplication object.
type CloudflareAccessApplicationReconciler struct {
	CloudflareAccessReconciler
	client.Client
	Scheme *runtime.Scheme
	Helper *ctrlhelper.ControllerHelper

	// Mainly used for debug / tests purposes. Should not be instantiated in production run.
	OptionalTracer *cfapi.InsertedCFRessourcesTracer
}

const (
	// StatusAvailable represents the status / condition type of the Cloudflare App.
	StatusAvailable = "Available"
)

// +kubebuilder:rbac:groups=cloudflare.zelic.io,resources=cloudflareaccessapplications,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cloudflare.zelic.io,resources=cloudflareaccessapplications/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=cloudflare.zelic.io,resources=cloudflareaccessapplications/finalizers,verbs=update

func (r *CloudflareAccessApplicationReconciler) GetReconcilierLogger(ctx context.Context) logr.Logger {
	return ctrl.LoggerFrom(ctx).WithName("CloudflareAccessApplicationController::Reconcile")
}

//nolint:maintidx,cyclop,gocognit,gocyclo,varnamelen
func (r *CloudflareAccessApplicationReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.GetReconcilierLogger(ctx)
	app := &v4alpha1.CloudflareAccessApplication{}
	var err error

	//
	// Try to get AccessApplication CRD Manifest
	//

	//
	if err = r.Get(ctx, req.NamespacedName, app); err != nil {
		// Not found ? might have been deleted; skip Reconciliation
		if k8serrors.IsNotFound(err) {
			return ctrl.Result{}, nil // will stop
		}

		// Else, return with failure
		return ctrl.Result{}, fault.Wrap(err, fmsg.With("Failed to get CloudflareAccessApplication")) // will retry immediately
	}

	//
	// Check that single app type are not declared elsewhere
	//

	//
	switch app.Spec.Type {
	case string(zero_trust.ApplicationTypeWARP),
		string(zero_trust.ApplicationTypeAppLauncher):
		{
			allApps := v4alpha1.CloudflareAccessApplicationList{}
			err = r.List(ctx, &allApps, client.MatchingFields{SearchOnAppTypeIndex: app.Spec.Type})
			if err != nil {
				// will retry immediately
				return ctrl.Result{RequeueAfter: r.Helper.NormalRequeueDelay}, fault.Wrap(err,
					fmsg.With("Failed to use indexed search on access applications. Contact the developers."),
					fctx.With(ctx,
						"searchedOn", SearchOnAppTypeIndex,
					),
				)
			}

			for _, existing := range allApps.Items {
				if existing.Namespace != app.Namespace || existing.Name != app.Name {
					//
					log.Error(
						fault.New(
							"Another unique-style access application definition already exist",
							fctx.With(ctx,
								"foundAsName", existing.Name,
								"foundInNamespace", existing.Namespace,
							),
						),
						"Having multiple definitions of an unique-style access application is forbidden",
					)

					// will stop
					return ctrl.Result{}, nil
				}
			}
		}
	}

	//
	// Ensure all PolicyRefs underlying [CloudflareAccessReusablePolicy] references exist and are available,
	// then store their CF IDs
	//

	//
	policyRefsNS, err := app.Spec.GetNamespacedPolicyRefs(ctx, req.Namespace)
	if err != nil {
		return ctrl.Result{}, fault.Wrap(err, fmsg.With("Failed to determine policy references")) // will retry immediately
	}

	//
	var arp v4alpha1.CloudflareAccessReusablePolicy //nolint:varnamelen
	orderedPolicyIds := []string{}
	for _, policyRefNS := range policyRefsNS {
		//
		err = r.Get(ctx, policyRefNS, &arp)
		if err != nil {
			// will retry immediately
			return ctrl.Result{}, fault.Wrap(err,
				fmsg.With("Reference to CloudflareAccessReusablePolicy do not exist"),
				fctx.With(ctx,
					"policyRef", v4alpha1.ParsedNamespacedName(policyRefNS),
				),
			)
		}

		//
		isAvailable := meta.IsStatusConditionPresentAndEqual(*arp.GetConditions(), StatusAvailable, metav1.ConditionTrue)
		if !isAvailable {
			log.Info("Referenced CloudflareAccessReusablePolicy not available yet, requeuing")
			return ctrl.Result{RequeueAfter: r.Helper.NormalRequeueDelay}, nil // will retry later
		}

		// if ready, we know for sure that AccessReusablePolicyID exists, so we extract it
		orderedPolicyIds = append(orderedPolicyIds, arp.Status.AccessReusablePolicyID)
	}

	app.Status.ReusablePolicyIDs = orderedPolicyIds
	if err = r.Client.Status().Update(ctx, app); err != nil {
		// will retry immediately
		return ctrl.Result{}, fault.Wrap(err, fmsg.With("Failed to update CloudflareAccessApplication status"))
	}

	//
	// Setup access to CF API
	//

	// Gather credentials to connect to Cloudflare's API
	cfConfig := config.ParseCloudflareConfig(app)
	validConfig, err := cfConfig.IsValid()
	if !validConfig {
		// will retry immediately
		return ctrl.Result{}, fault.Wrap(err, fmsg.With("invalid config"))
	}

	// Initialize Cloudflare's API wrapper
	api := cfapi.FromConfig(ctx, cfConfig, r.OptionalTracer)

	//
	// Proceed marked-as pending operations of manifest (if any)
	//

	// Attempt pending deletions on CRD Manifest
	continueReconcilliation, err := r.Helper.ReconcileDeletion(ctx, api, app)
	if !continueReconcilliation || err != nil {
		// will retry immediately
		return ctrl.Result{}, fault.Wrap(err, fmsg.With("unable to reconcile deletion for application"))
	}

	//
	// May mark manifest status state
	//

	// Try to setup "Conditions" field on CRD Manifest associated status
	_, err = controllerutil.CreateOrPatch(ctx, r.Client, app, func() error {
		meta.SetStatusCondition(&app.Status.Conditions,
			metav1.Condition{
				Type:    StatusAvailable,
				Status:  metav1.ConditionUnknown,
				Reason:  "Reconciling",
				Message: "CloudflareAccessApplication is reconciling",
			},
		)
		return nil
	})
	if err != nil {
		// will retry immediately
		return ctrl.Result{}, fault.Wrap(err, fmsg.With("Failed to update CloudflareAccessApplication status"))
	}

	//
	// Find CloudFlare Access Application depending on presence of CF UUID bound to resource
	//

	var cfAccessApp *zero_trust.AccessApplicationGetResponse

	//
	if app.GetCloudflareUUID() == "" {

		//
		// Has no UUID
		//

		switch app.Spec.Type {
		case string(zero_trust.ApplicationTypeSelfHosted):
			{
				cfAccessApp, err = api.FindAccessApplicationByDomain(ctx, app.Spec.Domain)
				if err != nil {
					// will retry immediately
					return ctrl.Result{}, fault.Wrap(err, fmsg.With("unable to get access application by domain"))
				}

				// else...

				//
				// ...if not found, we'll create it later then !
				//
			}
		case string(zero_trust.ApplicationTypeWARP),
			string(zero_trust.ApplicationTypeAppLauncher):
			{
				cfAccessApp, err = api.FindFirstAccessApplicationOfType(ctx, app.Spec.Type)

				if cfAccessApp == nil {
					//
					ctx := fctx.With(ctx,
						"advice", "Make sure you activated the associated feature in your cloudflare's dashboard !",
						"uniqueAppType", app.Spec.Type,
					)
					errMsg := "Issue finding unique app from cloudflare."

					// no API call error ? means it was not found
					if err == nil {
						//
						log.Error(fault.New("Missing application from Cloudflare", ctx), errMsg)

						// ... but since we cannot create it using CF API, just requeue until it has been activated by a user
						return ctrl.Result{RequeueAfter: r.Helper.NormalRequeueDelay}, nil
					}

					// API call error
					return ctrl.Result{}, fault.Wrap(
						err,
						fmsg.With(errMsg),
						ctx,
					)
				}
			}
		default:
			{
				// will retry immediately
				return ctrl.Result{}, fault.Newf("Unhandled application type '%s'. Contact the developers.", app.Spec.Type) //nolint:wrapcheck
			}
		}

		// would do nothing if resource was not found above
		if err = r.MayReconcileStatus(ctx, cfAccessApp, app); err != nil {
			// will retry immediately
			return ctrl.Result{}, fault.Wrap(err, fmsg.With("issue updating status"))
		}

	} else {

		//
		// Already has a UUID
		//

		// try to get the associated CF resource
		cfAccessApp, err = api.AccessApplication(ctx, app.GetCloudflareUUID())

		//
		if err != nil {
			// do not allow to continue if anything other than not found
			if !api.Is404(err) {
				// will retry immediately
				return ctrl.Result{}, fault.Wrap(err, fmsg.With("unable to get access application"))
			}

			// well, Application ID we had do not exist anymore; lets recreate the app in CF
			log.Info("access application ID linked to manifest not found - recreating remote resource...",
				"accessApplicationID", app.GetCloudflareUUID(),
			)

			// reset UUID so we can reconcile later on
			app.Status.AccessApplicationID = ""
		}
	}

	//
	// May create / recreate / update Access Application on CloudFlare API
	//

	if cfAccessApp == nil {

		//
		// no ressource found, create it with API
		//

		log.Info("app is missing - creating...",
			"name", app.Spec.Name,
			"domain", app.Spec.Domain,
		)
		cfAccessApp, err = api.CreateAccessApplication(ctx, app)
		if err != nil {
			// will retry immediately
			return ctrl.Result{}, fault.Wrap(err, fmsg.With("unable to create access application"))
		}

		log.Info("app successfully updated !")

		// update status
		if err = r.MayReconcileStatus(ctx, cfAccessApp, app); err != nil {
			// will retry immediately
			return ctrl.Result{}, fault.Wrap(err, fmsg.With("issue updating status"))
		}

	} else if needsUpdate := !cfcompare.AreAccessApplicationsEquivalent(ctx, &log, cfAccessApp, app); needsUpdate {

		//
		// diff found between fetched CF resource and definition
		//

		log.Info("app has changed - updating...",
			"name", app.Spec.Name,
			"domain", app.Spec.Domain,
		)

		//
		cfAccessApp, err = api.UpdateAccessApplication(ctx, app)
		if err != nil {
			// will retry immediately
			return ctrl.Result{}, fault.Wrap(err, fmsg.With("unable to update access group"))
		}

		log.Info("app successfully updated !")

		// update status
		if err = r.MayReconcileStatus(ctx, cfAccessApp, app); err != nil {
			// will retry immediately
			return ctrl.Result{}, fault.Wrap(err, fmsg.With("issue updating status"))
		}
	}

	//
	// All set, now mark ressource as available
	//

	if _, err = controllerutil.CreateOrPatch(ctx, r.Client, app, func() error {
		meta.SetStatusCondition(&app.Status.Conditions,
			metav1.Condition{
				Type:    StatusAvailable,
				Status:  metav1.ConditionTrue,
				Reason:  "Reconcilied",
				Message: "App Reconciled Successfully",
			},
		)

		return nil
	}); err != nil {
		// will retry immediately
		return ctrl.Result{}, fault.Wrap(err, fmsg.With("Failed to update CloudflareAccessApplication status"))
	}

	//
	// All good !
	//

	log.Info("changes successfully acknoledged")

	// will stop normally
	return ctrl.Result{}, nil
}

//nolint:dupl
func (r *CloudflareAccessApplicationReconciler) MayReconcileStatus(ctx context.Context, cfApp *zero_trust.AccessApplicationGetResponse, k8sApp *v4alpha1.CloudflareAccessApplication) error {
	if k8sApp.GetCloudflareUUID() != "" {
		return nil
	}
	if cfApp == nil {
		return nil
	}

	k8sApp.Status.AccessApplicationID = cfApp.ID
	k8sApp.Status.CreatedAt = metav1.NewTime(cfApp.CreatedAt)
	k8sApp.Status.UpdatedAt = metav1.NewTime(cfApp.UpdatedAt)

	if err := r.Client.Status().Update(ctx, k8sApp); err != nil {
		return fault.Wrap(err, fmsg.With("Failed to update CloudflareAccessApplication status"))
	}
	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *CloudflareAccessApplicationReconciler) SetupWithManager(mgr ctrl.Manager, override reconcile.Reconciler) error {
	if override == nil {
		override = r
	}

	ctx := context.TODO()

	//
	err := mgr.GetFieldIndexer().IndexField(ctx,
		&v4alpha1.CloudflareAccessApplication{},
		SearchOnAppTypeIndex,
		func(rawObj client.Object) []string {
			app := rawObj.(*v4alpha1.CloudflareAccessApplication)
			return []string{app.Spec.Type}
		},
	)

	//
	if err != nil {
		return fault.Wrap(err,
			fmsg.With("Unable to integrate custom index on access application controller. Contact the developers."),
			fctx.With(ctx,
				"index", SearchOnAppTypeIndex,
			),
		)
	}

	//nolint:wrapcheck
	return ctrl.NewControllerManagedBy(mgr).
		For(&v4alpha1.CloudflareAccessApplication{}, builder.WithPredicates(predicate.GenerationChangedPredicate{})).
		Complete(override)
}
