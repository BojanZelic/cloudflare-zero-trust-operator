//go:build e2e

/*
Copyright 2026.

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

package e2e

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/bojanzelic/cloudflare-zero-trust-operator/internal/cfapi"
	"github.com/bojanzelic/cloudflare-zero-trust-operator/internal/config"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const e2eNamespace = "cf-zero-trust-e2e"

var (
	cfClient    *cfapi.API
	managerCmd  *exec.Cmd
	managerDone chan error
	repoRoot    string
)

func TestE2E(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Cloudflare Zero Trust Operator E2E Suite")
}

func kubectl(stdin string, args ...string) (string, error) {
	fullArgs := append([]string{"--kuberc=/dev/null"}, args...)
	cmd := exec.Command("kubectl", fullArgs...)
	if stdin != "" {
		cmd.Stdin = strings.NewReader(stdin)
	}
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return stdout.String(), fmt.Errorf("kubectl %s failed: %w: %s", strings.Join(args, " "), err, stderr.String())
	}

	return strings.TrimSpace(stdout.String()), nil
}

var _ = BeforeSuite(func() {
	var err error
	repoRoot, err = filepath.Abs(filepath.Join("..", ".."))
	Expect(err).ToNot(HaveOccurred())

	By("Verifying kubectl connectivity to the local test cluster")
	nodesOut, err := kubectl("", "get", "nodes")
	Expect(err).ToNot(HaveOccurred(), "kubectl must be able to reach the local test cluster")
	GinkgoWriter.Printf("Connected to cluster nodes:\n%s\n", nodesOut)

	By("Initializing Cloudflare API client")
	config.SetConfigDefaults()
	cfCfg := config.ParseCloudflareConfig(&metav1.ObjectMeta{})
	valid, err := cfCfg.IsValid()
	Expect(err).ToNot(HaveOccurred())
	Expect(valid).To(BeTrue())

	cfClient, err = cfapi.New(cfCfg.APIToken, cfCfg.APIKey, cfCfg.APIEmail, cfCfg.AccountID)
	Expect(err).ToNot(HaveOccurred())

	By("Cleaning pre-existing Cloudflare test resources in account")
	ctx := context.Background()
	apps, err := cfClient.AccessApplications(ctx)
	Expect(err).ToNot(HaveOccurred())
	for _, app := range apps {
		Expect(cfClient.DeleteAccessApplication(ctx, app.ID)).To(Succeed())
	}
	groups, err := cfClient.AccessGroups(ctx)
	Expect(err).ToNot(HaveOccurred())
	for _, g := range groups {
		_ = cfClient.DeleteAccessGroup(ctx, g.ID)
	}
	tokens, err := cfClient.ServiceTokens(ctx)
	Expect(err).ToNot(HaveOccurred())
	for _, t := range tokens {
		_ = cfClient.DeleteAccessServiceToken(ctx, t.ID)
	}

	By("Installing CRDs onto the local cluster with kubectl")
	crdDir := filepath.Join(repoRoot, "config", "crd", "bases")
	_, err = kubectl("", "apply", "-f", crdDir)
	Expect(err).ToNot(HaveOccurred())

	_, err = kubectl("", "wait", "--for=condition=Established", "--timeout=30s",
		"crd/cloudflareaccessapplications.cloudflare.zelic.io",
		"crd/cloudflareaccessgroups.cloudflare.zelic.io",
		"crd/cloudflareservicetokens.cloudflare.zelic.io",
	)
	Expect(err).ToNot(HaveOccurred())

	By("Preparing E2E namespace on the local cluster")
	for _, resource := range []string{"cloudflareaccessapplications", "cloudflareaccessgroups", "cloudflareservicetokens"} {
		namesOut, listErr := kubectl("", "get", resource, "-n", e2eNamespace, "-o", "name", "--ignore-not-found=true")
		if listErr == nil && namesOut != "" {
			for _, objName := range strings.Split(namesOut, "\n") {
				objName = strings.TrimSpace(objName)
				if objName != "" {
					_, _ = kubectl("", "patch", objName, "-n", e2eNamespace, "--type=merge", "-p", `{"metadata":{"finalizers":null}}`)
				}
			}
		}
	}
	_, _ = kubectl("", "delete", "namespace", e2eNamespace, "--ignore-not-found=true", "--wait=true", "--timeout=30s")
	_, err = kubectl("", "create", "namespace", e2eNamespace)
	Expect(err).ToNot(HaveOccurred())

	By("Building the operator manager binary")
	managerBin := filepath.Join(repoRoot, "bin", "manager-e2e")
	buildCmd := exec.Command("go", "build", "-o", managerBin, "./cmd/main.go")
	buildCmd.Dir = repoRoot
	buildOut, err := buildCmd.CombinedOutput()
	Expect(err).ToNot(HaveOccurred(), string(buildOut))

	By("Starting the operator manager against the local kubectl cluster")
	managerCmd = exec.Command(managerBin,
		"--metrics-bind-address=0",
		"--health-probe-bind-address=127.0.0.1:18081",
	)
	managerCmd.Dir = repoRoot
	managerCmd.Env = os.Environ()
	managerCmd.Stdout = GinkgoWriter
	managerCmd.Stderr = GinkgoWriter
	Expect(managerCmd.Start()).To(Succeed())

	managerDone = make(chan error, 1)
	go func() {
		managerDone <- managerCmd.Wait()
	}()

	By("Waiting for operator manager /readyz probe")
	Eventually(func() error {
		resp, reqErr := http.Get("http://127.0.0.1:18081/readyz")
		if reqErr != nil {
			return reqErr
		}
		defer func() { _ = resp.Body.Close() }()
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("unexpected readyz status %d", resp.StatusCode)
		}
		return nil
	}, 20*time.Second, 500*time.Millisecond).Should(Succeed())
})

var _ = AfterSuite(func() {
	By("Cleaning up E2E custom resources and namespace via kubectl")
	_, _ = kubectl("", "delete", "cloudflareaccessapplications", "--all", "-n", e2eNamespace, "--ignore-not-found=true", "--timeout=60s")
	_, _ = kubectl("", "delete", "cloudflareaccessgroups", "--all", "-n", e2eNamespace, "--ignore-not-found=true", "--timeout=60s")
	_, _ = kubectl("", "delete", "cloudflareservicetokens", "--all", "-n", e2eNamespace, "--ignore-not-found=true", "--timeout=60s")
	_, _ = kubectl("", "delete", "namespace", e2eNamespace, "--ignore-not-found=true", "--wait=false")

	if managerCmd != nil && managerCmd.Process != nil {
		By("Stopping operator manager process")
		_ = managerCmd.Process.Signal(os.Interrupt)
		select {
		case <-managerDone:
		case <-time.After(5 * time.Second):
			_ = managerCmd.Process.Kill()
		}
	}

	_ = os.Remove(filepath.Join(repoRoot, "bin", "manager-e2e"))
})
