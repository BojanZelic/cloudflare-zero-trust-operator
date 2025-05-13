//go:build integration
// +build integration

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
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	cloudflarev4alpha1 "github.com/bojanzelic/cloudflare-zero-trust-operator/api/v4alpha1"
	"github.com/bojanzelic/cloudflare-zero-trust-operator/internal/cfapi"
	"github.com/bojanzelic/cloudflare-zero-trust-operator/internal/config"
	"github.com/bojanzelic/cloudflare-zero-trust-operator/internal/ctrlhelper"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
	//+kubebuilder:scaffold:imports
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var cfg *rest.Config
var k8sClient client.Client
var testEnv *envtest.Environment
var api *cfapi.API
var ctx context.Context
var cancel context.CancelFunc
var insertedTracer *cfapi.InsertedCFRessourcesTracer

//
//
//

// var logger logr.Logger
var logOutput *TestLogger

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecs(t, "Controller Suite")
}

var _ = BeforeSuite(func() {
	outLogger := zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true))
	logf.SetLogger(outLogger)

	ctx, cancel = context.WithCancel(context.TODO())

	By("bootstrapping test environment")
	testEnv = &envtest.Environment{
		CRDDirectoryPaths:     []string{filepath.Join("..", "..", "config", "crd", "bases")},
		ErrorIfCRDPathMissing: true,
	}

	var err error
	// cfg is defined in this file globally.
	cfg, err = testEnv.Start()
	Expect(err).NotTo(HaveOccurred())
	Expect(cfg).NotTo(BeNil())

	err = cloudflarev4alpha1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	//+kubebuilder:scaffold:scheme

	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
	Expect(err).NotTo(HaveOccurred())
	Expect(k8sClient).NotTo(BeNil())

	By("bootstrapping cloudflare api client")
	config.SetConfigDefaults()
	cfConfig := config.ParseCloudflareConfig(&v1.ObjectMeta{})
	_, err = cfConfig.IsValid()
	Expect(err).NotTo(HaveOccurred())

	insertedTracer = &cfapi.InsertedCFRessourcesTracer{}
	insertedTracer.ResetCFUUIDs()
	api = cfapi.New(cfConfig.APIToken, cfConfig.APIKey, cfConfig.APIEmail, cfConfig.AccountID, insertedTracer)

	logOutput = NewTestLogger(logr.RuntimeInfo{CallDepth: 1})

	By("bootstrapping managers")
	logger := logr.New(logOutput)
	ctrl.SetLogger(logger)

	k8sManager, err := ctrl.NewManager(cfg, ctrl.Options{
		Scheme:  scheme.Scheme,
		Metrics: metricsserver.Options{BindAddress: "0"},
		Logger:  logger,
	})
	Expect(err).ToNot(HaveOccurred())

	controllerHelper := &ctrlhelper.ControllerHelper{
		R: k8sClient,
	}

	Expect((&CloudflareAccessGroupReconciler{
		Client:         k8sClient,
		Scheme:         k8sClient.Scheme(),
		Helper:         controllerHelper,
		OptionalTracer: insertedTracer,
	}).SetupWithManager(k8sManager)).ToNot(HaveOccurred())
	Expect((&CloudflareAccessApplicationReconciler{
		Client:         k8sClient,
		Scheme:         k8sClient.Scheme(),
		Helper:         controllerHelper,
		OptionalTracer: insertedTracer,
	}).SetupWithManager(k8sManager)).ToNot(HaveOccurred())
	Expect((&CloudflareServiceTokenReconciler{
		Client:         k8sClient,
		Scheme:         k8sClient.Scheme(),
		Helper:         controllerHelper,
		OptionalTracer: insertedTracer,
	}).SetupWithManager(k8sManager)).ToNot(HaveOccurred())
	Expect((&CloudflareAccessReusablePolicyReconciler{
		Client:         k8sClient,
		Scheme:         k8sClient.Scheme(),
		Helper:         controllerHelper,
		OptionalTracer: insertedTracer,
	}).SetupWithManager(k8sManager)).ToNot(HaveOccurred())

	go func() {
		defer GinkgoRecover()

		err := k8sManager.Start(ctx)
		Expect(err).ToNot(HaveOccurred())

		err = testEnv.Stop()
		Expect(err).ToNot(HaveOccurred())
	}()
})

func NewTestLogger(info logr.RuntimeInfo) *TestLogger {
	return &TestLogger{
		Output: []map[string]any{},
		r:      info,
		mutex:  sync.RWMutex{},
	}
}

type TestLogger struct {
	Output []map[string]any
	r      logr.RuntimeInfo
	mutex  sync.RWMutex
}

func (t *TestLogger) doLog(level int, msg string, keysAndValues ...any) {
	t.mutex.Lock()
	m := map[string]any{}
	m["lvl"] = level
	m["keys"] = keysAndValues
	m["msg"] = msg
	m["time"] = time.Now()
	t.Output = append(t.Output, m)
	t.mutex.Unlock()
}

func (t *TestLogger) Clear() {
	t.mutex.Lock()
	t.Output = []map[string]any{}
	t.mutex.Unlock()
}

func (t *TestLogger) Init(info logr.RuntimeInfo) { t.r = info }
func (t *TestLogger) Enabled(level int) bool     { return true }
func (t *TestLogger) Info(level int, msg string, keysAndValues ...any) {
	t.doLog(level, msg, keysAndValues...)
}
func (t *TestLogger) Error(err error, msg string, keysAndValues ...any) {
	t.doLog(1, msg, append(keysAndValues, err)...)
}
func (t *TestLogger) WithValues(keysAndValues ...any) logr.LogSink { return t }
func (t *TestLogger) WithName(name string) logr.LogSink            { return t }
func (t *TestLogger) GetErrorCount() int {
	count := 0
	for _, out := range t.Output {
		if out["lvl"] == 1 {
			count++
		}
	}
	return count
}

func (t *TestLogger) GetOutput() string {
	typeMap := map[int]string{
		1: "ERROR",
		0: "INFO",
	}
	out := []string{}

	for _, val := range t.Output {
		t := val["time"].(time.Time)
		out = append(out, fmt.Sprintf("%v - %02d:%02d:%02d - %v - %v", typeMap[val["lvl"].(int)], t.Hour(), t.Minute(), t.Second(), val["msg"], val["keys"]))
	}

	return "\n" + strings.Join(out, "\n")
}
