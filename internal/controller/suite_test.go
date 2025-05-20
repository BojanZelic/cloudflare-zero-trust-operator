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

//
// Heavily inspired by https://github.com/kubernetes-sigs/kubebuilder/blob/master/docs/book/src/cronjob-tutorial/testdata/project/internal/controller/suite_test.go
//

package controller_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/viper"
	"go.uber.org/zap/zapcore"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
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

// @dev prevent direct access in tests
const defaultAccountOwnedDomain = "cf-operator-tests.uk"

var (
	k8sClient client.Client
	api       *cfapi.API
	testEnv   *envtest.Environment

	//
	insertedTracer cfapi.InsertedCFRessourcesTracer

	// @dev Might get cleared between test sets
	ctrlErrors controller.ReconcilierErrorTracker

	ctx    context.Context
	cancel context.CancelFunc

	// Domain which will be used as base for testing purposes (allowed email targets, allowed domains, access application domains...)
	//
	// This domain should be owned by provided CloudFlare account (by CLOUDFLARE_API_KEY / CLOUDFLARE_API_TOKEN / CLOUDFLARE_API_EMAIL)
	accountOwnedDomain string
)

const (
	defaultTimeout  = 10 * time.Second
	defaultPollRate = 2 * time.Second
)

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecs(t, "Controller Suite")
}

var _ = BeforeSuite(func() {

	ctx, cancel = context.WithCancel(context.TODO())

	//
	//
	//

	By("bootstrapping logger")
	// configure logger
	zapLogger := zap.New(
		zap.WriteTo(GinkgoWriter),
		zap.UseDevMode(true),
		zap.StacktraceLevel(zapcore.DPanicLevel), // only print stacktraces for panics and fatal
	)
	// bind logger
	ctrl.SetLogger(
		logger.NewFaultLogger(zapLogger,
			&logger.FaultLoggerOptions{
				DismissErrorVerbose: true,
			},
		),
	)

	//
	//
	//

	By("bootstrapping test environment")
	testEnv = &envtest.Environment{
		CRDDirectoryPaths:       []string{filepath.Join("..", "..", "config", "crd", "bases")},
		ErrorIfCRDPathMissing:   true,
		ControlPlaneStopTimeout: 1 * time.Minute, // Or any duration that suits your tests
	}

	if getFirstFoundEnvTestBinaryDir() != "" {
		testEnv.BinaryAssetsDirectory = getFirstFoundEnvTestBinaryDir()
	}

	cfg, err := testEnv.Start() // Required to be shutdown later
	Expect(err).NotTo(HaveOccurred())
	Expect(cfg).NotTo(BeNil())

	err = cloudflarev4alpha1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
	Expect(err).NotTo(HaveOccurred())
	Expect(k8sClient).NotTo(BeNil())

	//
	//
	//

	By("bootstrapping cloudflare api client")
	config.SetConfigDefaults()
	cfConfig := config.ParseCloudflareConfig(&v1.ObjectMeta{})
	_, err = cfConfig.IsValid()
	Expect(err).NotTo(HaveOccurred())

	insertedTracer.ResetCFUUIDs()
	api = cfapi.FromConfig(ctrl.SetupSignalHandler(), cfConfig, &insertedTracer)

	//
	// Picking testing domain
	//

	// Read from env
	accountOwnedDomain = viper.GetString("TEST_ACCOUNT_OWNED_DOMAIN")

	// if still empty...
	if accountOwnedDomain == "" {
		// resolves to default
		accountOwnedDomain = defaultAccountOwnedDomain

		//
		ctrl.Log.Info(
			"No account-owned domain picked from environment variable; resorting to default",
			"defaultedTestDomain", defaultAccountOwnedDomain,
		)
	} else {
		ctrl.Log.Info(
			"Using account-owned test domain, picked from environment",
			"testDomain", defaultAccountOwnedDomain,
		)
	}

	isOwned, err := api.IsDomainOwned(ctx, accountOwnedDomain)
	Expect(err).NotTo(HaveOccurred())
	Expect(isOwned).To(
		BeTrueBecause("Domain used as prop template during tests must be owned; inferred testing domain '%s' is not.", accountOwnedDomain),
	)

	//
	//
	//

	By("bootstrapping managers")
	k8sManager, err := ctrl.NewManager(cfg, ctrl.Options{
		Scheme:  scheme.Scheme,
		Metrics: metricsserver.Options{BindAddress: "0"},
	})
	Expect(err).ToNot(HaveOccurred())

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
	// start !
	//

	go func() {
		defer GinkgoRecover()
		err = k8sManager.Start(ctx)
		Expect(err).ToNot(HaveOccurred(), "failed to run manager")
	}()
})

var _ = AfterSuite(func() {
	By("tearing down the test environment")
	cancel()
	err := testEnv.Stop()
	Expect(err).NotTo(HaveOccurred())
})

//
//
//

// getFirstFoundEnvTestBinaryDir locates the first binary in the specified path.
// ENVTEST-based tests depend on specific binaries, usually located in paths set by
// controller-runtime. When running tests directly (e.g., via an IDE) without using
// Makefile targets, the 'BinaryAssetsDirectory' must be explicitly configured.
//
// This function streamlines the process by finding the required binaries, similar to
// setting the 'KUBEBUILDER_ASSETS' environment variable. To ensure the binaries are
// properly set up, run 'make setup-envtest' beforehand.
func getFirstFoundEnvTestBinaryDir() string {
	basePath := filepath.Join("..", "..", "bin", "k8s")
	entries, err := os.ReadDir(basePath)
	if err != nil {
		logf.Log.Error(err, "Failed to read directory", "path", basePath)
		return ""
	}
	for _, entry := range entries {
		if entry.IsDir() {
			return filepath.Join(basePath, entry.Name())
		}
	}
	return ""
}

//
//
//

// call this to update a CRD spec name, which should normally make it dirty by controllers
func addDirtyingSuffix(toUpdate *string) {
	if toUpdate != nil {
		*toUpdate = fmt.Sprintf("%s - updated name", *toUpdate)
	}
}

// Will generate an email address ([accountName]@<accountOwnedDomain>) that the configured CF account controls
func produceOwnedEmail(accountName string) string {
	return fmt.Sprintf("%s@%s", accountName, accountOwnedDomain)
}

// Will generate a Fully Qualified Domain name ([subdomain].<accountOwnedDomain>) that the configured CF account owns
func produceOwnedFQDN(subdomain string) string {
	return fmt.Sprintf("%s.%s", subdomain, accountOwnedDomain)
}

//
//
//

// @notice [res] will be populated
func ByExpectingCFResourceToBeReady(ctx context.Context, name types.NamespacedName, res ctrlhelper.CloudflareControlledResource) AsyncAssertion {
	return expectingCFResourceReadiness(ctx, name, res, metav1.ConditionTrue)
}

// @notice [res] will be populated
func ByExpectingCFResourceToNOTBeReady(ctx context.Context, name types.NamespacedName, res ctrlhelper.CloudflareControlledResource) AsyncAssertion {
	return expectingCFResourceReadiness(ctx, name, res, metav1.ConditionFalse)
}

// @notice [res] will be populated
func expectingCFResourceReadiness(
	ctx context.Context,
	name types.NamespacedName,
	res ctrlhelper.CloudflareControlledResource,
	awaitedCondition v1.ConditionStatus,
) AsyncAssertion {
	//
	const awaitedStatus = controller.StatusAvailable

	// check if [res] is already set, and if so; expect status update date to change while Eventually
	prevCond := meta.FindStatusCondition(*res.GetConditions(), awaitedStatus)

	//
	var byLabel string
	var prevTransTime time.Time
	if prevCond != nil {
		prevTransTime = prevCond.LastTransitionTime.Time
		byLabel = fmt.Sprintf(
			"Await for %s resource to process back to \"%s\"",
			res.Describe(),
			awaitedStatus,
		)
	} else {
		byLabel = fmt.Sprintf(
			"Await for %s resource to be \"%s\"",
			res.Describe(),
			awaitedStatus,
		)
	}

	//
	//
	//

	//
	By(byLabel)

	//
	return Eventually(func(g Gomega) { //nolint:varnamelen
		//
		err := k8sClient.Get(ctx, name, res)
		g.Expect(err).To(Not(HaveOccurred()))

		//
		g.Expect(res.GetCloudflareUUID()).ToNot(BeEmpty())

		//
		cond := meta.FindStatusCondition(*res.GetConditions(), awaitedStatus)
		g.Expect(cond).NotTo(BeNil())
		g.Expect(cond.Status).To(Equal(awaitedCondition))

		//
		if !prevTransTime.IsZero() {
			g.Expect(cond.LastTransitionTime.Time).To(BeTemporally(">", prevTransTime))
		}

	}).WithTimeout(defaultTimeout).WithPolling(defaultPollRate)
}
