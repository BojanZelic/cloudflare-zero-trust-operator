// TODO: add back //go:build integration

package controller_test

import (
	"context"

	v4alpha1 "github.com/bojanzelic/cloudflare-zero-trust-operator/api/v4alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

const updtdName = "updated name"

var _ = Describe("CloudflareAccessApplication controller", Ordered, func() {
	BeforeAll(func() { insertedTracer.ResetCFUUIDs() })
	AfterAll(func() { insertedTracer.UninstallFromCF(api) })

	//
	//
	//

	Context("CloudflareAccessApplication controller test", func() {
		const testScopedNamespace = "zto-testing-app"

		//
		ctx := context.Background()
		testNS := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: testScopedNamespace,
			},
		}

		BeforeEach(func() {
			ctrlErrors.Clear()

			By("Creating the Namespace to perform the tests")
			_ = k8sClient.Create(ctx, testNS)
			// ignore error because of https://book.kubebuilder.io/reference/envtest.html#namespace-usage-limitation
			// Expect(err).To(Not(HaveOccurred()))
		})

		AfterEach(func() {
			// By("expect no reconcile errors occurred")
			// Expect(ctrlErrors).To(BeEmpty())
			By("Deleting the Namespace to perform the tests")
			_ = k8sClient.Delete(ctx, testNS)
		})

		//
		//
		//

		It("should successfully reconcile CloudflareAccessApplication with reusable policy", func() {
			By("Creating the custom resource for the Kind CloudflareAccessReusablePolicy")
			arpNN := types.NamespacedName{Name: "test-1-arp", Namespace: testScopedNamespace}
			reusablePolicy := &v4alpha1.CloudflareAccessReusablePolicy{
				ObjectMeta: metav1.ObjectMeta{
					Name:      arpNN.Name,
					Namespace: arpNN.Namespace,
				},
				Spec: v4alpha1.CloudflareAccessReusablePolicySpec{
					Name: "ZTO AccessApplication Tests - 1 - Policy",
					Include: v4alpha1.CloudFlareAccessRules{
						Emails:       []string{"testemail@cf-operator-tests.uk", "testemail2@cf-operator-tests.uk"},
						EmailDomains: []string{"cf-operator-tests.uk"},
					},
				},
			}
			Expect(k8sClient.Create(ctx, reusablePolicy)).To(Not(HaveOccurred()))

			foundArp := &v4alpha1.CloudflareAccessReusablePolicy{}
			ByExpectingCFResourceToBeReady(ctx,
				arpNN,
				foundArp,
			)

			//
			//
			//

			By("Creating the custom resource for the Kind CloudflareAccessApplication")
			appNN := types.NamespacedName{Name: "test-1-app", Namespace: testScopedNamespace}
			app := &v4alpha1.CloudflareAccessApplication{
				ObjectMeta: metav1.ObjectMeta{
					Name:      appNN.Name,
					Namespace: appNN.Namespace,
				},
				Spec: v4alpha1.CloudflareAccessApplicationSpec{
					Name:   "ZTO AccessApplication Tests - 1 - App",
					Domain: "integration-policies.cf-operator-tests.uk",
					PolicyRefs: []string{
						v4alpha1.ParsedNamespacedName(appNN),
					},
				},
			}
			Expect(k8sClient.Create(ctx, app)).To(Not(HaveOccurred()))

			foundApp := &v4alpha1.CloudflareAccessApplication{}
			ByExpectingCFResourceToBeReady(ctx,
				appNN,
				foundApp,
			)

			//
			//
			//

			By("Ensuring Cloudflare Application refers Reusable policy CF ID")
			Expect(foundApp.Status.ReusablePolicyIDs).To(ContainElement(foundArp.GetCloudflareUUID()))
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
					Domain: "integration.cf-operator-tests.uk",
				},
			}
			Expect(k8sClient.Create(ctx, app)).To(Not(HaveOccurred()))

			//
			foundApp := &v4alpha1.CloudflareAccessApplication{}
			ByExpectingCFResourceToBeReady(ctx,
				appNN,
				foundApp,
			)

			By("Cloudflare resource should equal the spec")
			cfResource, err := api.AccessApplication(ctx, foundApp.GetCloudflareUUID())
			Expect(err).To(Not(HaveOccurred()))
			Expect(cfResource.Name).To(Equal(foundApp.Spec.Name))

			By("Updating app name")
			foundApp.Spec.Name = updtdName
			Expect(k8sClient.Update(ctx, foundApp)).To(Not(HaveOccurred()))

			// Await for resource to be ready again
			ByExpectingCFResourceToBeReady(ctx,
				appNN,
				foundApp,
			)

			By("Cloudflare resource should equal the updated spec")
			Eventually(func(g Gomega) { //nolint:varnamelen
				// ctrlErrors.TestEmpty()
				cfResource, err = api.AccessApplication(ctx, foundApp.GetCloudflareUUID())
				g.Expect(err).To(Not(HaveOccurred()))
				g.Expect(cfResource.Name).To(Equal(foundApp.Spec.Name))
			}).WithTimeout(defaultTimeout).WithPolling(defaultPoolRate).Should(Succeed(), ctrlErrors) // sometimes this is cached

			By("Cloudflare resource should be deleted")
			Expect(k8sClient.Delete(ctx, app)).To(Not(HaveOccurred()))

			By("Checking if the custom resource was successfully deleted")
			Eventually(func() error {
				// ctrlErrors.TestEmpty()
				return k8sClient.Get(ctx, appNN, app)
			}).WithTimeout(defaultTimeout).WithPolling(defaultPoolRate).Should(Not(Succeed()))
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
					Domain:  "integration-logo-test.cf-operator-tests.uk",
					LogoURL: "https://www.cloudflare.com/img/logo-web-badges/cf-logo-on-white-bg.svg",
				},
			}
			Expect(k8sClient.Create(ctx, app)).To(Not(HaveOccurred()))

			//
			foundApp := &v4alpha1.CloudflareAccessApplication{}
			ByExpectingCFResourceToBeReady(ctx,
				appNN,
				foundApp,
			)

			By("Cloudflare resource should equal the spec")
			cfResource, err := api.AccessApplication(ctx, foundApp.GetCloudflareUUID())
			Expect(err).To(Not(HaveOccurred()))
			Expect(cfResource.Name).To(Equal(foundApp.Spec.Name))
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
					Domain: "recreate-application.cf-operator-tests.uk",
				},
			}
			Expect(k8sClient.Create(ctx, app)).To(Not(HaveOccurred()))

			//
			foundApp := &v4alpha1.CloudflareAccessApplication{}
			ByExpectingCFResourceToBeReady(ctx,
				appNN,
				foundApp,
			)

			By("Delete associated CF Application")
			oldAccessApplicationID := foundApp.GetCloudflareUUID()
			Expect(api.DeleteAccessApplication(ctx, foundApp.GetCloudflareUUID())).To(Not(HaveOccurred()))

			By("re-trigger reconcile by updating access application")
			foundApp.Spec.Name = updtdName
			Expect(k8sClient.Update(ctx, foundApp)).To(Not(HaveOccurred()))

			//
			ByExpectingCFResourceToBeReady(ctx,
				appNN,
				foundApp,
			)
			Expect(foundApp.GetCloudflareUUID()).ToNot(Equal(oldAccessApplicationID))

			By("Cloudflare resource should equal the updated spec")
			cfResource, err := api.AccessApplication(ctx, foundApp.GetCloudflareUUID())
			Expect(err).To(Not(HaveOccurred()))
			Expect(cfResource.Name).To(Equal(foundApp.Spec.Name))
		})
	})
})
