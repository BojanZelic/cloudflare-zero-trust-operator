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

	// initialize client which we will use in test, emulating operator / remote client
	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
	Expect(err).NotTo(HaveOccurred())
	Expect(k8sClient).NotTo(BeNil())

	//
	//
	//

	By("bootstrapping cloudflare api client")
	config.SetConfigDefaults()
	cfConfig := config.ParseCloudflareConfig(&metav1.ObjectMeta{})
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

	/*
		One thing that this autogenerated file is missing, however, is a way to actually start your controller.
		The code above will set up a client for interacting with your custom Kind,
		but will not be able to test your controller behavior.
		If you want to test your custom controller logic, you’ll need to add some familiar-looking manager logic
		to your BeforeSuite() function, so you can register your custom controller to run on this test cluster.

		You may notice that the code below runs your controller with nearly identical logic to your CronJob project’s main.go!
		The only difference is that the manager is started in a separate goroutine so it does not block the cleanup of envtest
		when you’re done running your tests.

		Note that we set up both a "live" k8s client and a separate client from the manager. This is because when making
		assertions in tests, you generally want to assert against the live state of the API server. If you use the client
		from the manager (`k8sManager.GetClient`), you'd end up asserting against the contents of the cache instead, which is
		slower and can introduce flakiness into your tests. We could use the manager's `APIReader` to accomplish the same
		thing, but that would leave us with two clients in our test assertions and setup (one for reading, one for writing),
		and it'd be easy to make mistakes.

		Note that we keep the reconciler running against the manager's cache client, though -- we want our controller to
		behave as it would in production, and we use features of the cache (like indices) in our controller which aren't
		available when talking directly to the API server.
	*/
	By("bootstrapping managers")
	k8sManager, err := ctrl.NewManager(cfg, ctrl.Options{
		Scheme:  scheme.Scheme,
		Metrics: metricsserver.Options{BindAddress: "0"},
	})
	Expect(err).ToNot(HaveOccurred())

	controllerHelper := &ctrlhelper.ControllerHelper{
		R:                  k8sManager.GetClient(),
		NormalRequeueDelay: 2 * time.Second,
	}
	ctrlErrors = controller.ReconcilierErrorTracker{}

	Expect((&controller.ReconcilerWithLoggedErrors{
		ErrTracker: &ctrlErrors,
		Inner: &controller.CloudflareAccessGroupReconciler{
			Client:         k8sManager.GetClient(),
			Scheme:         k8sManager.GetScheme(),
			Helper:         controllerHelper,
			OptionalTracer: &insertedTracer,
		},
	}).SetupWithManager(k8sManager)).ToNot(HaveOccurred())
	Expect((&controller.ReconcilerWithLoggedErrors{
		ErrTracker: &ctrlErrors,
		Inner: &controller.CloudflareAccessApplicationReconciler{
			Client:         k8sManager.GetClient(),
			Scheme:         k8sManager.GetScheme(),
			Helper:         controllerHelper,
			OptionalTracer: &insertedTracer,
		},
	}).SetupWithManager(k8sManager)).ToNot(HaveOccurred())
	Expect((&controller.ReconcilerWithLoggedErrors{
		ErrTracker: &ctrlErrors,
		Inner: &controller.CloudflareServiceTokenReconciler{
			Client:         k8sManager.GetClient(),
			Scheme:         k8sManager.GetScheme(),
			Helper:         controllerHelper,
			OptionalTracer: &insertedTracer,
		},
	}).SetupWithManager(k8sManager)).ToNot(HaveOccurred())
	Expect((&controller.ReconcilerWithLoggedErrors{
		ErrTracker: &ctrlErrors,
		Inner: &controller.CloudflareAccessReusablePolicyReconciler{
			Client:         k8sManager.GetClient(),
			Scheme:         k8sManager.GetScheme(),
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
func ByExpectingCFResourceToBeReady(ctx context.Context, toExpectOf ctrlhelper.CloudflareControlledResource) AsyncAssertion {
	return expectingCFResourceReadiness(ctx, toExpectOf, metav1.ConditionTrue)
}

// @notice [res] will be populated
func ByExpectingCFResourceToNOTBeReady(ctx context.Context, toExpectOf ctrlhelper.CloudflareControlledResource) AsyncAssertion {
	return expectingCFResourceReadiness(ctx, toExpectOf, metav1.ConditionFalse)
}

// @notice [res] will be populated
func expectingCFResourceReadiness(
	ctx context.Context,
	toExpectOf ctrlhelper.CloudflareControlledResource,
	awaitedCondition metav1.ConditionStatus,
) AsyncAssertion {
	//
	const awaitedStatus = controller.StatusAvailable

	// check if [res] is already set, and if so; expect status update date to change while Eventually
	prevCond := meta.FindStatusCondition(*toExpectOf.GetConditions(), awaitedStatus)

	//
	var byLabel string
	var prevTransTime time.Time
	if prevCond != nil {
		prevTransTime = prevCond.LastTransitionTime.Time
		byLabel = fmt.Sprintf(
			"Await for %s resource to process back to \"%s\"",
			toExpectOf.Describe(),
			awaitedStatus,
		)
	} else {
		byLabel = fmt.Sprintf(
			"Await for %s resource to be \"%s\"",
			toExpectOf.Describe(),
			awaitedStatus,
		)
	}

	//
	//
	//

	//
	By(byLabel)

	//
	toExpectOfNN := toExpectOf.GetNamespacedName()

	//
	return Eventually(func(g Gomega) { //nolint:varnamelen
		//
		err := k8sClient.Get(ctx, toExpectOfNN, toExpectOf)
		g.Expect(err).To(Not(HaveOccurred()))

		//
		g.Expect(toExpectOf.GetCloudflareUUID()).ToNot(BeEmpty())

		//
		cond := meta.FindStatusCondition(*toExpectOf.GetConditions(), awaitedStatus)
		g.Expect(cond).NotTo(BeNil())
		g.Expect(cond.Status).To(Equal(awaitedCondition))

		//
		if !prevTransTime.IsZero() {
			g.Expect(cond.LastTransitionTime.Time).To(BeTemporally(">", prevTransTime))
		}

	}).WithTimeout(defaultTimeout).WithPolling(defaultPollRate)
}
