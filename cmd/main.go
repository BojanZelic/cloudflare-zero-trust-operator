/*
Copyright 2024.

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

package main

import (
	"context"
	"crypto/tls"
	"flag"
	"os"
	"time"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, SAML, etc.)
	// to ensure that exec-entrypoint and run can make use of them.

	_ "k8s.io/client-go/plugin/pkg/client/auth"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/metrics/filters"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	cloudflarev4 "github.com/bojanzelic/cloudflare-zero-trust-operator/api/v4alpha1"
	"github.com/bojanzelic/cloudflare-zero-trust-operator/internal/cfapi"
	"github.com/bojanzelic/cloudflare-zero-trust-operator/internal/config"
	"github.com/bojanzelic/cloudflare-zero-trust-operator/internal/controller"
	"github.com/bojanzelic/cloudflare-zero-trust-operator/internal/ctrlhelper"
	// +kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(cloudflarev4.AddToScheme(scheme))
	// +kubebuilder:scaffold:scheme
}

//nolint:cyclop
func main() {
	var metricsAddr string
	var enableLeaderElection bool
	var probeAddr string
	var secureMetrics bool
	var enableHTTP2 bool
	var tlsOpts []func(*tls.Config)
	flag.StringVar(&metricsAddr, "metrics-bind-address", "0", "The address the metrics endpoint binds to. "+
		"Use :8443 for HTTPS or :8080 for HTTP, or leave as 0 to disable the metrics service.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	flag.BoolVar(&secureMetrics, "metrics-secure", true,
		"If set, the metrics endpoint is served securely via HTTPS. Use --metrics-secure=false to use HTTP instead.")
	flag.BoolVar(&enableHTTP2, "enable-http2", false,
		"If set, HTTP/2 will be enabled for the metrics and webhook servers")

	// Configure logger
	opts := zap.Options{
		// TODO missing fctx, customize encoder ?
		Development: true,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()
	rootLogger := zap.New(
		zap.UseFlagOptions(&opts),
	)
	ctrl.SetLogger(rootLogger)

	// if the enable-http2 flag is false (the default), http/2 should be disabled
	// due to its vulnerabilities. More specifically, disabling http/2 will
	// prevent from being vulnerable to the HTTP/2 Stream Cancellation and
	// Rapid Reset CVEs. For more information see:
	// - https://github.com/advisories/GHSA-qppj-fm5r-hxr3
	// - https://github.com/advisories/GHSA-4374-p667-p6c8
	disableHTTP2 := func(c *tls.Config) {
		setupLog.Info("disabling http/2")
		c.NextProtos = []string{"http/1.1"}
	}

	if !enableHTTP2 {
		tlsOpts = append(tlsOpts, disableHTTP2)
	}

	webhookServer := webhook.NewServer(webhook.Options{
		TLSOpts: tlsOpts,
	})

	// Metrics endpoint is enabled in 'config/default/kustomization.yaml'. The Metrics options configure the server.
	// More info:
	// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.19.1/pkg/metrics/server
	// - https://book.kubebuilder.io/reference/metrics.html
	metricsServerOptions := metricsserver.Options{
		BindAddress:   metricsAddr,
		SecureServing: secureMetrics,
		TLSOpts:       tlsOpts,
	}

	if secureMetrics {
		// FilterProvider is used to protect the metrics endpoint with authn/authz.
		// These configurations ensure that only authorized users and service accounts
		// can access the metrics endpoint. The RBAC are configured in 'config/rbac/kustomization.yaml'. More info:
		// https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.19.1/pkg/metrics/filters#WithAuthenticationAndAuthorization
		metricsServerOptions.FilterProvider = filters.WithAuthenticationAndAuthorization

		// TODO(maintainer): If CertDir, CertName, and KeyName are not specified, controller-runtime will automatically
		// generate self-signed certificates for the metrics server. While convenient for development and testing,
		// this setup is not recommended for production.
	}

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		Metrics:                metricsServerOptions,
		WebhookServer:          webhookServer,
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "8cdce734.zelic.io",
		// LeaderElectionReleaseOnCancel defines if the leader should step down voluntarily
		// when the Manager ends. This requires the binary to immediately end when the
		// Manager is stopped, otherwise, this setting is unsafe. Setting this significantly
		// speeds up voluntary leader transitions as the new leader don't have to wait
		// LeaseDuration time first.
		//
		// In the default scaffold provided, the program ends immediately after
		// the manager stops, so would be fine to enable this option. However,
		// if you are doing or is intended to do any operation such as perform cleanups
		// after the manager stops then its usage might be unsafe.
		// LeaderElectionReleaseOnCancel: true,
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	config.SetConfigDefaults()
	displayAvailableIdentityProviders()

	controllerHelper := &ctrlhelper.ControllerHelper{
		R:                  mgr.GetClient(),
		NormalRequeueDelay: 10 * time.Second,
	}

	if err = (&controller.ReconcilerWithLoggedErrors{
		Inner: &controller.CloudflareAccessReusablePolicyReconciler{
			Client:         mgr.GetClient(),
			Scheme:         mgr.GetScheme(),
			Helper:         controllerHelper,
			OptionalTracer: nil,
		},
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller",
			"controller", "CloudflareAccessReusablePolicy",
		)
		os.Exit(1)
	}
	if err = (&controller.ReconcilerWithLoggedErrors{
		Inner: &controller.CloudflareAccessGroupReconciler{
			Client:         mgr.GetClient(),
			Scheme:         mgr.GetScheme(),
			Helper:         controllerHelper,
			OptionalTracer: nil,
		},
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller",
			"controller", "CloudflareAccessGroup",
		)
		os.Exit(1)
	}
	if err = (&controller.ReconcilerWithLoggedErrors{
		Inner: &controller.CloudflareServiceTokenReconciler{
			Client:         mgr.GetClient(),
			Scheme:         mgr.GetScheme(),
			Helper:         controllerHelper,
			OptionalTracer: nil,
		},
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller",
			"controller", "CloudflareServiceToken",
		)
		os.Exit(1)
	}
	if err = (&controller.ReconcilerWithLoggedErrors{
		Inner: &controller.CloudflareAccessApplicationReconciler{
			Client:         mgr.GetClient(),
			Scheme:         mgr.GetScheme(),
			Helper:         controllerHelper,
			OptionalTracer: nil,
		},
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller",
			"controller", "CloudflareAccessApplication",
		)
		os.Exit(1)
	}
	// +kubebuilder:scaffold:builder

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}

// show in the logs what Identity providers are available
func displayAvailableIdentityProviders() {
	//
	idpsUsedIn := []string{
		"CloudflareAccessApplication.Spec.allowedIdps",
		"CloudflareAccessReusablePolicy.Spec.{include,exclude,require}.loginMethods[]",
		"CloudflareAccessReusablePolicy.Spec.{include,exclude,require}.googleGroups[].identityProviderId",
		"CloudflareAccessReusablePolicy.Spec.{include,exclude,require}.oktaGroups[].identityProviderId",
		"CloudflareAccessReusablePolicy.Spec.{include,exclude,require}.samlGroups[].identityProviderId",
		"CloudflareAccessReusablePolicy.Spec.{include,exclude,require}.githubOrganizations[].identityProviderId",
	}

	//
	setupLog.Info("Checking available Identity Providers "+
		"that you may use to configure this operator...",
		"IdentityProvidersUsedIn",
		idpsUsedIn,
	)

	// Gather credentials to connect to Cloudflare's API
	cfConfig := config.ParseCloudflareConfig(&metav1.ObjectMeta{})
	validConfig, err := cfConfig.IsValid()
	if !validConfig {
		setupLog.Error(err, "invalid config")
		os.Exit(1)
	}

	// Initialize Cloudflare's API wrapper
	ctx := context.TODO()
	api := cfapi.FromConfig(ctx, cfConfig, nil)
	idProviders, err := api.IdentityProviders(ctx)
	if err != nil {
		setupLog.Error(err, "failed to fetch env account identity providers")
		os.Exit(1)
	}

	if len(*idProviders) == 0 {
		setupLog.Info(
			//
			"No identity providers found; "+
				"you might want to enable some through CloudFlare's dashboard "+
				"to leverage most of this operator's features.",
			//
			"moreInfosAt", "https://developers.cloudflare.com/cloudflare-one/identity/",
		)
		return
	}

	//
	setupLog.Info(
		//
		"Enumerating found identity providers; "+
			"please use their UUID as reference within this operator, as listed below.",
		//
		"AvailableIDPs", len(*idProviders),
	)

	for i, idProvider := range *idProviders {
		setupLog.Info("Found IdentityProvider",
			//
			"order", i,
			"type", idProvider.Type,
			"name", idProvider.Name,
			"uuid", idProvider.ID,
		)
	}
}

//
//
//
