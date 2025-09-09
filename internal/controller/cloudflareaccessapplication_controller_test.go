//go:build integration

package controller_test

import (
	"context"
	"strings"

	v4alpha1 "github.com/bojanzelic/cloudflare-zero-trust-operator/api/v4alpha1"
	"github.com/cloudflare/cloudflare-go/v4/zero_trust"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("CloudflareAccessApplication controller", Ordered, func() {
	BeforeAll(func() { insertedTracer.ResetStores() })
	AfterAll(func() {
		errs := insertedTracer.UninstallFromCF(api)
		Expect(errs).To(BeEmpty())
	})

	//
	//
	//

	const testScopedNamespace = "zto-testing-app"

	//
	ctx := context.Background()
	testNS := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: testScopedNamespace,
		},
	}

	BeforeAll(func() {
		By("Creating the Namespace to perform the tests")
		_ = k8sClient.Create(ctx, testNS)
	})
	AfterAll(func() {
		By("Deleting the Namespace to perform the tests")
		_ = k8sClient.Delete(ctx, testNS)
		// ignore error because of https://book.kubebuilder.io/reference/envtest.html#namespace-usage-limitation
		// Expect(err).ToNot(HaveOccurred()))
	})

	BeforeEach(func() {
		ctrlErrors.Clear()
	})
	AfterEach(func() {
		// By("expect no reconcile errors occurred")
		// Expect(ctrlErrors).To(BeEmpty())
	})

	//
	// Only Self hosted
	//

	Context("CloudflareAccessApplication controller test - self hosted apps", func() {
		It("should successfully reconcile CloudflareAccessApplication with reusable policy", func() {
			By("Creating the custom resource for the Kind CloudflareAccessReusablePolicy")
			arpNN := types.NamespacedName{Name: "test-1-arp", Namespace: testScopedNamespace}
			arp := &v4alpha1.CloudflareAccessReusablePolicy{
				ObjectMeta: metav1.ObjectMeta{
					Name:      arpNN.Name,
					Namespace: arpNN.Namespace,
				},
				Spec: v4alpha1.CloudflareAccessReusablePolicySpec{
					Name: "ZTO AccessApplication Tests - 1 - Policy",
					Include: v4alpha1.CloudFlareAccessRules{
						Emails: []string{
							produceOwnedEmail("zto-test-app-1-one"),
							produceOwnedEmail("zto-test-app-1-two"),
						},
						EmailDomains: []string{
							accountOwnedDomain,
						},
					},
				},
			}
			Expect(k8sClient.Create(ctx, arp)).ToNot(HaveOccurred())

			//
			ByExpectingCFResourceToBeReady(ctx, arp).Should(Succeed())

			By("Creating the custom resource for the Kind CloudflareAccessApplication")
			appNN := types.NamespacedName{Name: "test-1-app", Namespace: testScopedNamespace}
			app := &v4alpha1.CloudflareAccessApplication{
				ObjectMeta: metav1.ObjectMeta{
					Name:      appNN.Name,
					Namespace: appNN.Namespace,
				},
				Spec: v4alpha1.CloudflareAccessApplicationSpec{
					Name:   "ZTO AccessApplication Tests - 1 - App",
					Domain: produceOwnedFQDN("zto-test-app-1"),
					PolicyRefs: []string{
						v4alpha1.ParsedNamespacedName(arpNN),
					},
				},
			}
			Expect(k8sClient.Create(ctx, app)).ToNot(HaveOccurred())

			//
			ByExpectingCFResourceToBeReady(ctx, app).Should(Succeed())

			//
			//
			//

			By("Ensuring Cloudflare Application refers Reusable policy CF ID")
			Expect(app.Status.ReusablePolicyIDs).To(ContainElement(arp.GetCloudflareUUID()))
		})

		It("should successfully reconcile and delete a custom resource for CloudflareAccessApplication", func() {
			By("Creating the custom resource for the Kind CloudflareAccessApplication")
			appNN := types.NamespacedName{Name: "test-2-app", Namespace: testScopedNamespace}
			app := &v4alpha1.CloudflareAccessApplication{
				ObjectMeta: metav1.ObjectMeta{
					Name:      appNN.Name,
					Namespace: appNN.Namespace,
				},
				Spec: v4alpha1.CloudflareAccessApplicationSpec{
					Name:   "ZTO AccessApplication Tests - 2 - App",
					Domain: produceOwnedFQDN("zto-test-app-2"),
				},
			}
			Expect(k8sClient.Create(ctx, app)).ToNot(HaveOccurred())

			//
			ByExpectingCFResourceToBeReady(ctx, app).Should(Succeed())

			By("Cloudflare resource should equal the spec")
			cfResource, err := api.AccessApplication(ctx, app.GetCloudflareUUID())
			Expect(err).ToNot(HaveOccurred())
			Expect(cfResource.Name).To(Equal(app.Spec.Name))

			By("Updating app name")
			addDirtyingSuffix(&app.Spec.Name)
			Expect(k8sClient.Update(ctx, app)).ToNot(HaveOccurred())

			// Await for resource to be ready again
			ByExpectingCFResourceToBeReady(ctx, app).Should(Succeed())

			By("Cloudflare resource should equal the updated spec")
			Eventually(func(g Gomega) { //nolint:varnamelen
				// ctrlErrors.TestEmpty()
				cfResource, err = api.AccessApplication(ctx, app.GetCloudflareUUID())
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(cfResource.Name).To(Equal(app.Spec.Name))
			}).WithTimeout(defaultTimeout).WithPolling(defaultPollRate).Should(Succeed(), ctrlErrors) // sometimes this is cached

			By("Deleting cloudflare resource")
			Expect(k8sClient.Delete(ctx, app)).ToNot(HaveOccurred())

			//
			ByExpectingDeletionOf(app).Should(Succeed())
		})

		It("should be able to set a LogoURL for CloudflareAccessApplication", func() {
			By("Creating the custom resource for the Kind CloudflareAccessApplication")
			appNN := types.NamespacedName{Name: "test-3-app", Namespace: testScopedNamespace}
			app := &v4alpha1.CloudflareAccessApplication{
				ObjectMeta: metav1.ObjectMeta{
					Name:      appNN.Name,
					Namespace: appNN.Namespace,
				},
				Spec: v4alpha1.CloudflareAccessApplicationSpec{
					Name:    "ZTO AccessApplication Tests - 3 - App",
					Domain:  produceOwnedFQDN("zto-test-app-3"),
					LogoURL: "https://www.cloudflare.com/img/logo-web-badges/cf-logo-on-white-bg.svg",
				},
			}
			Expect(k8sClient.Create(ctx, app)).ToNot(HaveOccurred())

			//
			ByExpectingCFResourceToBeReady(ctx, app).Should(Succeed())

			By("Cloudflare resource should equal the spec")
			cfResource, err := api.AccessApplication(ctx, app.GetCloudflareUUID())
			Expect(err).ToNot(HaveOccurred())
			Expect(cfResource.Name).To(Equal(app.Spec.Name))
		})

		It("should successfully reconcile CloudflareAccessApplication whose AccessApplicationID references a missing Application", func() {
			By("Recreating the custom resource for the Kind CloudflareAccessApplication")
			appNN := types.NamespacedName{Name: "test-4-app", Namespace: testScopedNamespace}
			app := &v4alpha1.CloudflareAccessApplication{
				ObjectMeta: metav1.ObjectMeta{
					Name:      appNN.Name,
					Namespace: appNN.Namespace,
				},
				Spec: v4alpha1.CloudflareAccessApplicationSpec{
					Name:   "ZTO AccessApplication Tests - 4 - App",
					Domain: produceOwnedFQDN("zto-test-app-4"),
				},
			}
			Expect(k8sClient.Create(ctx, app)).ToNot(HaveOccurred())

			//
			ByExpectingCFResourceToBeReady(ctx, app).Should(Succeed())

			By("Delete associated CF Application")
			oldAccessApplicationID := app.GetCloudflareUUID()
			Expect(api.DeleteOrResetAccessApplication(ctx, app)).ToNot(HaveOccurred())

			By("re-trigger reconcile by updating access application")
			addDirtyingSuffix(&app.Spec.Name)
			Expect(k8sClient.Update(ctx, app)).ToNot(HaveOccurred())

			// Await for resource to be ready again
			ByExpectingCFResourceToBeReady(ctx, app).Should(Succeed())
			Expect(app.GetCloudflareUUID()).ToNot(Equal(oldAccessApplicationID))

			By("Cloudflare resource should equal the updated spec")
			cfResource, err := api.AccessApplication(ctx, app.GetCloudflareUUID())
			Expect(err).ToNot(HaveOccurred())
			Expect(cfResource.Name).To(Equal(app.Spec.Name))
		})
	})

	//
	// only one-per-account app types
	//

	Context("CloudflareAccessApplication controller test - one per account apps", func() {

		const setKey = "SingleApp:"

		var cfAppThen *zero_trust.AccessApplicationGetResponse
		var err error

		//
		produceLabelFor := func(appType zero_trust.ApplicationType) Labels {
			return Label(setKey + string(appType))
		}

		//
		extractAppTypeFromLabel := func() string {
			for _, label := range CurrentSpecReport().Labels() {
				after, found := strings.CutPrefix(label, setKey)
				if found {
					return after
				}
			}
			panic("Unable to determine app type from label")
		}

		BeforeEach(func() {
			By("Backing up existing policy UUIDs")
			oneTimeAppType := extractAppTypeFromLabel()
			cfAppThen, err = api.FindFirstAccessApplicationOfType(ctx, oneTimeAppType)
			Expect(err).ToNot(HaveOccurred())
			Expect(cfAppThen).ToNot(BeNil())
		})

		AfterEach(func() {
			By("Restore existing policy UUIDs")
			err = api.RestoreAccessApplicationTo(ctx, cfAppThen)
			Expect(err).ToNot(HaveOccurred())
		})

		//
		//
		//

		It("Manage WARP Access Application", produceLabelFor(zero_trust.ApplicationTypeWARP), func() {
			oneTimeAppType := extractAppTypeFromLabel()

			By("Creating the custom resource for the Kind CloudflareAccessReusablePolicy")
			arpNN := types.NamespacedName{Name: "test-5-arp", Namespace: testScopedNamespace}
			reusablePolicy := &v4alpha1.CloudflareAccessReusablePolicy{
				ObjectMeta: metav1.ObjectMeta{
					Name:      arpNN.Name,
					Namespace: arpNN.Namespace,
				},
				Spec: v4alpha1.CloudflareAccessReusablePolicySpec{
					Name: "ZTO AccessApplication Tests - 5 - Policy",
					Include: v4alpha1.CloudFlareAccessRules{
						Emails: []string{
							produceOwnedEmail("zto-test-app-5-one"),
							produceOwnedEmail("zto-test-app-5-two"),
						},
						EmailDomains: []string{
							accountOwnedDomain,
						},
					},
				},
			}
			Expect(k8sClient.Create(ctx, reusablePolicy)).ToNot(HaveOccurred())

			//
			ByExpectingCFResourceToBeReady(ctx, reusablePolicy).Should(Succeed())

			By("Creating a WARP CloudflareAccessApplication")
			appNN := types.NamespacedName{Name: "test-5-app-warp", Namespace: testScopedNamespace}
			app := &v4alpha1.CloudflareAccessApplication{
				ObjectMeta: metav1.ObjectMeta{
					Name:      appNN.Name,
					Namespace: appNN.Namespace,
				},
				Spec: v4alpha1.CloudflareAccessApplicationSpec{
					Type: oneTimeAppType,
					PolicyRefs: []string{
						v4alpha1.ParsedNamespacedName(arpNN),
					},
				},
			}
			Expect(k8sClient.Create(ctx, app)).ToNot(HaveOccurred())

			//
			ByExpectingCFResourceToBeReady(ctx, app).Should(Succeed())

			//
			// OK ! now, reset
			//

			By("Deleting (resetting) the WARP resource")
			Expect(k8sClient.Delete(ctx, app)).ToNot(HaveOccurred())

			//
			ByExpectingDeletionOf(app).Should(Succeed())
		})

		It("Manage App Launcher Access Application", produceLabelFor(zero_trust.ApplicationTypeAppLauncher), func() {
			oneTimeAppType := extractAppTypeFromLabel()

			By("Creating the custom resource for the Kind CloudflareAccessReusablePolicy")
			arpNN := types.NamespacedName{Name: "test-6-arp", Namespace: testScopedNamespace}
			reusablePolicy := &v4alpha1.CloudflareAccessReusablePolicy{
				ObjectMeta: metav1.ObjectMeta{
					Name:      arpNN.Name,
					Namespace: arpNN.Namespace,
				},
				Spec: v4alpha1.CloudflareAccessReusablePolicySpec{
					Name: "ZTO AccessApplication Tests - 6 - Policy",
					Include: v4alpha1.CloudFlareAccessRules{
						Emails: []string{
							produceOwnedEmail("zto-test-app-6-one"),
							produceOwnedEmail("zto-test-app-6-two"),
						},
						EmailDomains: []string{
							accountOwnedDomain,
						},
					},
				},
			}
			Expect(k8sClient.Create(ctx, reusablePolicy)).ToNot(HaveOccurred())

			//
			ByExpectingCFResourceToBeReady(ctx, reusablePolicy).Should(Succeed())

			By("Creating an App Launcher CloudflareAccessApplication")
			appNN := types.NamespacedName{Name: "test-6-app-app-launcher", Namespace: testScopedNamespace}
			app := &v4alpha1.CloudflareAccessApplication{
				ObjectMeta: metav1.ObjectMeta{
					Name:      appNN.Name,
					Namespace: appNN.Namespace,
				},
				Spec: v4alpha1.CloudflareAccessApplicationSpec{
					Type: oneTimeAppType,
					PolicyRefs: []string{
						v4alpha1.ParsedNamespacedName(arpNN),
					},
				},
			}
			Expect(k8sClient.Create(ctx, app)).ToNot(HaveOccurred())

			//
			ByExpectingCFResourceToBeReady(ctx, app).Should(Succeed())

			//
			// OK ! now, reset
			//

			By("Deleting (resetting) the App Launcher resource")
			Expect(k8sClient.Delete(ctx, app)).ToNot(HaveOccurred())

			//
			ByExpectingDeletionOf(app).Should(Succeed())
		})
	})
})
