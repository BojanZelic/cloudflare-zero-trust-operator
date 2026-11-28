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
	"context"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/bojanzelic/cloudflare-zero-trust-operator/internal/cfapi"
	"github.com/bojanzelic/cloudflare-zero-trust-operator/internal/cftypes"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Cloudflare Zero Trust Operator E2E", Ordered, func() {
	ctx := context.Background()

	It("migrates pre-existing legacy Cloudflare resources and inline policies seamlessly without CRD changes", func() {
		By("Pre-creating a legacy Access Group and Access Application with an inline (non-reusable) policy in Cloudflare")
		preGroup, err := cfClient.CreateAccessGroup(ctx, cftypes.AccessGroup{
			Name: "e2e-legacy-group",
			Include: []interface{}{
				cfapi.NewAccessGroupEmail("legacy-group@cf-operator-tests.uk"),
			},
		})
		Expect(err).ToNot(HaveOccurred())

		tTrue := true
		tFalse := false
		domain := "e2e-legacy-migration.cf-operator-tests.uk"
		preApp, err := cfClient.CreateAccessApplication(ctx, cftypes.AccessApplication{
			Name:                    "e2e legacy app",
			Domain:                  domain,
			Type:                    cftypes.AccessApplicationTypeSelfHosted,
			AppLauncherVisible:      &tTrue,
			AutoRedirectToIdentity:  &tFalse,
			EnableBindingCookie:     &tFalse,
			HttpOnlyCookieAttribute: &tTrue,
			SessionDuration:         "24h",
		})
		Expect(err).ToNot(HaveOccurred())

		legacyPolID, err := cfClient.CreateLegacyApplicationPolicy(ctx, preApp.ID, cftypes.AccessPolicy{
			Name:       "e2e-legacy-inline-policy",
			Decision:   "allow",
			Precedence: 1,
			Include: []interface{}{
				cfapi.NewAccessGroupEmail("legacy-app@cf-operator-tests.uk"),
			},
		})
		Expect(err).ToNot(HaveOccurred())
		Expect(legacyPolID).ToNot(BeEmpty())

		polsBefore, err := cfClient.AccessPolicies(ctx, preApp.ID)
		Expect(err).ToNot(HaveOccurred())
		Expect(polsBefore).To(HaveLen(1))
		Expect(polsBefore[0].Reusable).To(BeFalse(), "pre-existing policy must start as legacy non-reusable")

		By("Applying existing v1alpha1 manifests via kubectl")
		manifests := fmt.Sprintf(`
apiVersion: cloudflare.zelic.io/v1alpha1
kind: CloudflareAccessGroup
metadata:
  name: e2e-legacy-group
  namespace: %s
spec:
  name: e2e-legacy-group
  include:
    - emails:
        - legacy-group@cf-operator-tests.uk
---
apiVersion: cloudflare.zelic.io/v1alpha1
kind: CloudflareAccessApplication
metadata:
  name: e2e-legacy-app
  namespace: %s
spec:
  name: e2e legacy app
  domain: %s
  policies:
    - name: e2e-legacy-inline-policy
      decision: allow
      include:
        - emails:
            - legacy-app@cf-operator-tests.uk
`, e2eNamespace, e2eNamespace, domain)

		_, err = kubectl(manifests, "apply", "-f", "-")
		Expect(err).ToNot(HaveOccurred())

		By("Waiting via kubectl for both resources to become Available and preserve Cloudflare IDs")
		_, err = kubectl("", "wait", "--for=condition=Available", "--timeout=45s",
			"cloudflareaccessgroup/e2e-legacy-group",
			"cloudflareaccessapplication/e2e-legacy-app",
			"-n", e2eNamespace,
		)
		Expect(err).ToNot(HaveOccurred())

		gotGroupID, err := kubectl("", "get", "cloudflareaccessgroup", "e2e-legacy-group", "-n", e2eNamespace, "-o", "jsonpath={.status.accessGroupId}")
		Expect(err).ToNot(HaveOccurred())
		Expect(gotGroupID).To(Equal(preGroup.ID))

		gotAppID, err := kubectl("", "get", "cloudflareaccessapplication", "e2e-legacy-app", "-n", e2eNamespace, "-o", "jsonpath={.status.accessApplicationId}")
		Expect(err).ToNot(HaveOccurred())
		Expect(gotAppID).To(Equal(preApp.ID))

		By("Verifying the legacy policy was migrated to a Cloudflare Reusable Policy")
		Eventually(func(g Gomega) {
			polsAfter, pErr := cfClient.AccessPolicies(ctx, preApp.ID)
			g.Expect(pErr).ToNot(HaveOccurred())
			g.Expect(polsAfter).To(HaveLen(1))
			g.Expect(polsAfter[0].Reusable).To(BeTrue(), "policy should be migrated to reusable")
			g.Expect(polsAfter[0].Name).To(Equal("e2e-legacy-inline-policy"))
		}, 30*time.Second, time.Second).Should(Succeed())

		By("Updating the migrated CloudflareAccessApplication via kubectl apply")
		updatedAppManifest := fmt.Sprintf(`
apiVersion: cloudflare.zelic.io/v1alpha1
kind: CloudflareAccessApplication
metadata:
  name: e2e-legacy-app
  namespace: %s
spec:
  name: e2e migrated app updated
  domain: %s
  policies:
    - name: e2e-reusable-policy-updated
      decision: allow
      include:
        - emails:
            - legacy-app@cf-operator-tests.uk
            - second-user@cf-operator-tests.uk
`, e2eNamespace, domain)

		_, err = kubectl(updatedAppManifest, "apply", "-f", "-")
		Expect(err).ToNot(HaveOccurred())

		Eventually(func(g Gomega) {
			cfApp, aErr := cfClient.AccessApplication(ctx, preApp.ID)
			g.Expect(aErr).ToNot(HaveOccurred())
			g.Expect(cfApp.Name).To(Equal("e2e migrated app updated"))

			polsUpdated, pErr := cfClient.AccessPolicies(ctx, preApp.ID)
			g.Expect(pErr).ToNot(HaveOccurred())
			g.Expect(polsUpdated).To(HaveLen(1))
			g.Expect(polsUpdated[0].Reusable).To(BeTrue())
			g.Expect(polsUpdated[0].Name).To(Equal("e2e-reusable-policy-updated"))
			g.Expect(polsUpdated[0].Include).To(HaveLen(2))
		}, 45*time.Second, time.Second).Should(Succeed())

		By("Deleting migrated resources via kubectl and verifying cleanup in Cloudflare")
		_, err = kubectl("", "delete", "cloudflareaccessapplication", "e2e-legacy-app", "-n", e2eNamespace, "--timeout=45s")
		Expect(err).ToNot(HaveOccurred())
		_, err = kubectl("", "delete", "cloudflareaccessgroup", "e2e-legacy-group", "-n", e2eNamespace, "--timeout=45s")
		Expect(err).ToNot(HaveOccurred())

		Eventually(func(g Gomega) {
			_, aErr := cfClient.AccessApplication(ctx, preApp.ID)
			g.Expect(cfapi.IsNotFoundError(aErr)).To(BeTrue(), "expected AccessApplication NotFound, got %v", aErr)
			_, gErr := cfClient.AccessGroup(ctx, preGroup.ID)
			g.Expect(cfapi.IsNotFoundError(gErr)).To(BeTrue(), "expected AccessGroup NotFound, got %v", gErr)
		}, 15*time.Second, time.Second).Should(Succeed())
	})

	It("creates, updates, cross-references, and deletes ServiceToken, AccessGroup, and AccessApplication via kubectl", func() {
		By("Creating CloudflareServiceToken and CloudflareAccessGroup via kubectl")
		prereqManifests := fmt.Sprintf(`
apiVersion: cloudflare.zelic.io/v1alpha1
kind: CloudflareServiceToken
metadata:
  name: e2e-service-token
  namespace: %s
spec:
  name: e2e service token
  template:
    metadata:
      name: e2e-service-token-secret
---
apiVersion: cloudflare.zelic.io/v1alpha1
kind: CloudflareAccessGroup
metadata:
  name: e2e-access-group
  namespace: %s
spec:
  name: e2e access group
  include:
    - emails:
        - team@cf-operator-tests.uk
    - ipRanges:
        - 203.0.113.10/32
`, e2eNamespace, e2eNamespace)

		_, err := kubectl(prereqManifests, "apply", "-f", "-")
		Expect(err).ToNot(HaveOccurred())

		_, err = kubectl("", "wait", "--for=condition=Available", "--timeout=45s",
			"cloudflareservicetoken/e2e-service-token",
			"cloudflareaccessgroup/e2e-access-group",
			"-n", e2eNamespace,
		)
		Expect(err).ToNot(HaveOccurred())

		tokenID, err := kubectl("", "get", "cloudflareservicetoken", "e2e-service-token", "-n", e2eNamespace, "-o", "jsonpath={.status.serviceTokenId}")
		Expect(err).ToNot(HaveOccurred())
		Expect(tokenID).ToNot(BeEmpty())

		By("Verifying the generated Kubernetes Secret contains valid ClientID and ClientSecret")
		clientIDB64, err := kubectl("", "get", "secret", "e2e-service-token-secret", "-n", e2eNamespace, "-o", "jsonpath={.data.cloudflareClientId}")
		Expect(err).ToNot(HaveOccurred())
		clientIDBytes, err := base64.StdEncoding.DecodeString(clientIDB64)
		Expect(err).ToNot(HaveOccurred())
		Expect(string(clientIDBytes)).To(ContainSubstring(".access"))

		clientSecretB64, err := kubectl("", "get", "secret", "e2e-service-token-secret", "-n", e2eNamespace, "-o", "jsonpath={.data.cloudflareSecretKey}")
		Expect(err).ToNot(HaveOccurred())
		clientSecretBytes, err := base64.StdEncoding.DecodeString(clientSecretB64)
		Expect(err).ToNot(HaveOccurred())
		Expect(string(clientSecretBytes)).ToNot(BeEmpty())

		groupID, err := kubectl("", "get", "cloudflareaccessgroup", "e2e-access-group", "-n", e2eNamespace, "-o", "jsonpath={.status.accessGroupId}")
		Expect(err).ToNot(HaveOccurred())
		Expect(groupID).ToNot(BeEmpty())

		By("Creating CloudflareAccessApplication referencing both the AccessGroup and ServiceToken via kubectl")
		appDomain := "e2e-full-workflow.cf-operator-tests.uk"
		appManifest := fmt.Sprintf(`
apiVersion: cloudflare.zelic.io/v1alpha1
kind: CloudflareAccessApplication
metadata:
  name: e2e-full-app
  namespace: %s
spec:
  name: e2e full workflow app
  domain: %s
  sessionDuration: 12h
  logoUrl: https://www.cloudflare.com/img/logo-web-badges/cf-logo-on-white-bg.svg
  policies:
    - name: Allow E2E Group
      decision: allow
      include:
        - accessGroups:
            - valueFrom:
                name: e2e-access-group
                namespace: %s
    - name: Allow E2E Service Token
      decision: non_identity
      include:
        - serviceToken:
            - valueFrom:
                name: e2e-service-token
                namespace: %s
`, e2eNamespace, appDomain, e2eNamespace, e2eNamespace)

		_, err = kubectl(appManifest, "apply", "-f", "-")
		Expect(err).ToNot(HaveOccurred())

		_, err = kubectl("", "wait", "--for=condition=Available", "--timeout=45s",
			"cloudflareaccessapplication/e2e-full-app",
			"-n", e2eNamespace,
		)
		Expect(err).ToNot(HaveOccurred())

		appID, err := kubectl("", "get", "cloudflareaccessapplication", "e2e-full-app", "-n", e2eNamespace, "-o", "jsonpath={.status.accessApplicationId}")
		Expect(err).ToNot(HaveOccurred())
		Expect(appID).ToNot(BeEmpty())

		By("Verifying Cloudflare state matches all 3 Kubernetes CRs")
		cfApp, err := cfClient.AccessApplication(ctx, appID)
		Expect(err).ToNot(HaveOccurred())
		Expect(cfApp.Name).To(Equal("e2e full workflow app"))
		Expect(cfApp.SessionDuration).To(Equal("12h"))

		cfPolicies, err := cfClient.AccessPolicies(ctx, appID)
		Expect(err).ToNot(HaveOccurred())
		Expect(cfPolicies).To(HaveLen(2))
		Expect(cfPolicies[0].Reusable).To(BeTrue())
		Expect(cfPolicies[0].Name).To(Equal("Allow E2E Group"))
		Expect(cfPolicies[1].Reusable).To(BeTrue())
		Expect(cfPolicies[1].Name).To(Equal("Allow E2E Service Token"))
		firstPolicyID := cfPolicies[0].ID
		secondPolicyID := cfPolicies[1].ID

		By("Updating ServiceToken, AccessGroup, and AccessApplication via kubectl")
		updateManifests := fmt.Sprintf(`
apiVersion: cloudflare.zelic.io/v1alpha1
kind: CloudflareServiceToken
metadata:
  name: e2e-service-token
  namespace: %s
spec:
  name: e2e service token renamed
  template:
    metadata:
      name: e2e-service-token-secret
---
apiVersion: cloudflare.zelic.io/v1alpha1
kind: CloudflareAccessGroup
metadata:
  name: e2e-access-group
  namespace: %s
spec:
  name: e2e access group renamed
  include:
    - emails:
        - team@cf-operator-tests.uk
        - admin@cf-operator-tests.uk
---
apiVersion: cloudflare.zelic.io/v1alpha1
kind: CloudflareAccessApplication
metadata:
  name: e2e-full-app
  namespace: %s
spec:
  name: e2e full workflow app renamed
  domain: %s
  sessionDuration: 24h
  policies:
    - name: Allow E2E Group Updated
      decision: allow
      include:
        - accessGroups:
            - valueFrom:
                name: e2e-access-group
                namespace: %s
`, e2eNamespace, e2eNamespace, e2eNamespace, appDomain, e2eNamespace)

		_, err = kubectl(updateManifests, "apply", "-f", "-")
		Expect(err).ToNot(HaveOccurred())

		Eventually(func(g Gomega) {
			updatedGroup, gErr := cfClient.AccessGroup(ctx, groupID)
			g.Expect(gErr).ToNot(HaveOccurred())
			g.Expect(updatedGroup.Name).To(Equal("e2e access group renamed"))
			g.Expect(updatedGroup.Include).To(HaveLen(2))

			tokens, tErr := cfClient.ServiceTokens(ctx)
			g.Expect(tErr).ToNot(HaveOccurred())
			foundRenamed := false
			for _, t := range tokens {
				if t.ID == tokenID && t.Name == "e2e service token renamed" {
					foundRenamed = true
				}
			}
			g.Expect(foundRenamed).To(BeTrue())

			updatedApp, aErr := cfClient.AccessApplication(ctx, appID)
			g.Expect(aErr).ToNot(HaveOccurred())
			g.Expect(updatedApp.Name).To(Equal("e2e full workflow app renamed"))
			g.Expect(updatedApp.SessionDuration).To(Equal("24h"))

			updatedPols, pErr := cfClient.AccessPolicies(ctx, appID)
			g.Expect(pErr).ToNot(HaveOccurred())
			g.Expect(updatedPols).To(HaveLen(1))
			g.Expect(updatedPols[0].Name).To(Equal("Allow E2E Group Updated"))

			_, removedPolErr := cfClient.GetAccessPolicy(ctx, secondPolicyID)
			g.Expect(cfapi.IsNotFoundError(removedPolErr)).To(BeTrue(), "removed reusable policy should be deleted from Cloudflare")
		}, 45*time.Second, time.Second).Should(Succeed())

		By("Deleting all 3 CRs via kubectl and verifying complete cleanup in Cloudflare")
		_, err = kubectl("", "delete", "cloudflareaccessapplication", "e2e-full-app", "-n", e2eNamespace, "--timeout=45s")
		Expect(err).ToNot(HaveOccurred())
		_, err = kubectl("", "delete", "cloudflareaccessgroup", "e2e-access-group", "-n", e2eNamespace, "--timeout=45s")
		Expect(err).ToNot(HaveOccurred())
		_, err = kubectl("", "delete", "cloudflareservicetoken", "e2e-service-token", "-n", e2eNamespace, "--timeout=45s")
		Expect(err).ToNot(HaveOccurred())

		Eventually(func(g Gomega) {
			_, aErr := cfClient.AccessApplication(ctx, appID)
			g.Expect(cfapi.IsNotFoundError(aErr)).To(BeTrue(), "expected AccessApplication NotFound, got %v", aErr)
			_, pErr := cfClient.GetAccessPolicy(ctx, firstPolicyID)
			g.Expect(cfapi.IsNotFoundError(pErr)).To(BeTrue(), "expected AccessPolicy NotFound, got %v", pErr)
			_, gErr := cfClient.AccessGroup(ctx, groupID)
			g.Expect(cfapi.IsNotFoundError(gErr)).To(BeTrue(), "expected AccessGroup NotFound, got %v", gErr)
		}, 15*time.Second, time.Second).Should(Succeed())
	})
})
