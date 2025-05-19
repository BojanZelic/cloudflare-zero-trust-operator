// TODO: add back //go:build integration

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

package controller_test

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/zap/zapcore"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	cloudflarev4alpha1 "github.com/bojanzelic/cloudflare-zero-trust-operator/api/v4alpha1"
	"github.com/bojanzelic/cloudflare-zero-trust-operator/internal/cfapi"
	"github.com/bojanzelic/cloudflare-zero-trust-operator/internal/config"
	"github.com/bojanzelic/cloudflare-zero-trust-operator/internal/controller"
	"github.com/bojanzelic/cloudflare-zero-trust-operator/internal/ctrlhelper"
	"github.com/bojanzelic/cloudflare-zero-trust-operator/internal/logger"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var k8sClient client.Client
var api *cfapi.API

var insertedTracer cfapi.InsertedCFRessourcesTracer

// @dev Might get cleared between test sets
var ctrlErrors controller.ReconcilierErrorTracker

const (
	defaultTimeout  = 10 * time.Second
	defaultPoolRate = 2 * time.Second
)

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecs(t, "Controller Suite")
}

var _ = BeforeSuite(func() {
	// configure logger
	zapLogger := zap.New(
		zap.WriteTo(GinkgoWriter),
		zap.UseDevMode(true),
		zap.StacktraceLevel(zapcore.DPanicLevel), // only print stacktraces for panics and fatal
	)

	// bind logger
	ctrl.SetLogger(
		logger.NewFaultLogger(zapLogger, &logger.FaultLoggerOptions{
			DismissErrorVerbose: true,
		}),
	)

	By("bootstrapping test environment")
	testEnv := &envtest.Environment{
		CRDDirectoryPaths:     []string{filepath.Join("..", "..", "config", "crd", "bases")},
		ErrorIfCRDPathMissing: true,
	}

	// cfg is defined in this file globally.
	cfg, err := testEnv.Start()
	Expect(err).NotTo(HaveOccurred())
	Expect(cfg).NotTo(BeNil())

	err = cloudflarev4alpha1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
	Expect(err).NotTo(HaveOccurred())
	Expect(k8sClient).NotTo(BeNil())

	By("bootstrapping cloudflare api client")
	config.SetConfigDefaults()
	cfConfig := config.ParseCloudflareConfig(&v1.ObjectMeta{})
	_, err = cfConfig.IsValid()
	Expect(err).NotTo(HaveOccurred())

	insertedTracer.ResetCFUUIDs()
	ctx := context.TODO()
	api = cfapi.FromConfig(ctx, cfConfig, &insertedTracer)

	By("bootstrapping managers")
	k8sManager, err := ctrl.NewManager(cfg, ctrl.Options{
		Scheme:  scheme.Scheme,
		Metrics: metricsserver.Options{BindAddress: "0"},
	})
	Expect(err).ToNot(HaveOccurred())

	//
	//
	//

	controllerHelper := &ctrlhelper.ControllerHelper{
		R:                  k8sClient,
		NormalRequeueDelay: 2 * time.Second,
	}
	ctrlErrors = controller.ReconcilierErrorTracker{}

	Expect((&controller.ReconcilerWithLoggedErrors{
		ErrTracker: &ctrlErrors,
		Inner: &controller.CloudflareAccessGroupReconciler{
			Client:         k8sClient,
			Scheme:         k8sClient.Scheme(),
			Helper:         controllerHelper,
			OptionalTracer: &insertedTracer,
		},
	}).SetupWithManager(k8sManager)).ToNot(HaveOccurred())
	Expect((&controller.ReconcilerWithLoggedErrors{
		ErrTracker: &ctrlErrors,
		Inner: &controller.CloudflareAccessApplicationReconciler{
			Client:         k8sClient,
			Scheme:         k8sClient.Scheme(),
			Helper:         controllerHelper,
			OptionalTracer: &insertedTracer,
		},
	}).SetupWithManager(k8sManager)).ToNot(HaveOccurred())
	Expect((&controller.ReconcilerWithLoggedErrors{
		ErrTracker: &ctrlErrors,
		Inner: &controller.CloudflareServiceTokenReconciler{
			Client:         k8sClient,
			Scheme:         k8sClient.Scheme(),
			Helper:         controllerHelper,
			OptionalTracer: &insertedTracer,
		},
	}).SetupWithManager(k8sManager)).ToNot(HaveOccurred())
	Expect((&controller.ReconcilerWithLoggedErrors{
		ErrTracker: &ctrlErrors,
		Inner: &controller.CloudflareAccessReusablePolicyReconciler{
			Client:         k8sClient,
			Scheme:         k8sClient.Scheme(),
			Helper:         controllerHelper,
			OptionalTracer: &insertedTracer,
		},
	}).SetupWithManager(k8sManager)).ToNot(HaveOccurred())

	//
	//
	//

	//
	go func() {
		defer GinkgoRecover()

		//
		err := k8sManager.Start(ctx)
		Expect(err).ToNot(HaveOccurred())

		//
		err = testEnv.Stop()
		Expect(err).ToNot(HaveOccurred())
	}()
})

//
//
//

// @notice [res] will be populated
func ByExpectingCFResourceToBeReady(ctx context.Context, name types.NamespacedName, res ctrlhelper.CloudflareControlledResource) {
	expectingCFResourceReadiness(ctx, name, res, metav1.ConditionTrue)
}

// @notice [res] will be populated
func ByExpectingCFResourceToNOTBeReady(ctx context.Context, name types.NamespacedName, res ctrlhelper.CloudflareControlledResource) {
	expectingCFResourceReadiness(ctx, name, res, metav1.ConditionFalse)
}

// @notice [res] will be populated
func expectingCFResourceReadiness(
	ctx context.Context,
	name types.NamespacedName,
	res ctrlhelper.CloudflareControlledResource,
	awaitedCondition v1.ConditionStatus,
) {
	//
	const awaitedStatus = controller.StatusAvailable

	//
	By(fmt.Sprintf("Await for %s resource to be \"%s\"",
		res.GetObjectKind().GroupVersionKind().Kind,
		awaitedStatus,
	))

	//
	Eventually(func(g Gomega) { //nolint:varnamelen
		//
		err := k8sClient.Get(ctx, name, res)
		g.Expect(err).To(Not(HaveOccurred()))

		//
		g.Expect(res.GetCloudflareUUID()).ToNot(BeEmpty())

		//
		conditions := res.GetConditions()
		matchesAwaitedStatus := meta.IsStatusConditionPresentAndEqual(*conditions, awaitedStatus, awaitedCondition)
		g.Expect(matchesAwaitedStatus).To(BeTrue())

	}).WithTimeout(defaultTimeout).WithPolling(defaultPoolRate).Should(Succeed())
}
