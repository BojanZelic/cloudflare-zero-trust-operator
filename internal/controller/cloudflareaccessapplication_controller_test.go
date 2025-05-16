// TODO: add back //go:build integration

package controller

import (
	"context"
	"time"

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

		const cloudflareName = "cloudflare-app"

		ctx := context.Background()

		namespace := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name:      cloudflareName,
				Namespace: cloudflareName,
			},
		}

		BeforeEach(func() {
			ctrlErrors.Clear()

			By("Creating the Namespace to perform the tests")
			_ = k8sClient.Create(ctx, namespace)
			// ignore error because of https://book.kubebuilder.io/reference/envtest.html#namespace-usage-limitation
			// Expect(err).To(Not(HaveOccurred()))
		})

		AfterEach(func() {
			By("expect no reconcile errors occurred")
			// Expect(ctrlErrors).To(BeEmpty())
			// 	By("Deleting the Namespace to perform the tests")
			// 	//_ = k8sClient.Delete(ctx, namespace)
		})

		//
		//
		//

		It("should successfully reconcile CloudflareAccessApplication with reusable policy", func() {
			By("Creating the custom resource for the Kind CloudflareAccessReusablePolicy")
			RPTypeNamespaceName := types.NamespacedName{Name: "integration-test-manifest", Namespace: cloudflareName}

			reusablePolicy := &v4alpha1.CloudflareAccessReusablePolicy{
				ObjectMeta: metav1.ObjectMeta{
					Name:      RPTypeNamespaceName.Name,
					Namespace: namespace.Name,
				},
				Spec: v4alpha1.CloudflareAccessReusablePolicySpec{
					Name: "integration_test",
					Include: v4alpha1.CloudFlareAccessRules{
						Emails:       []string{"testemail@cf-operator-tests.uk", "testemail2@cf-operator-tests.uk"},
						EmailDomains: []string{"cf-operator-tests.uk"},
					},
				},
			}

			err := k8sClient.Create(ctx, reusablePolicy)
			Expect(err).To(Not(HaveOccurred()))

			RPFound := &v4alpha1.CloudflareAccessReusablePolicy{}
			By("Checking the latest Status should have the ID of the resource")
			Eventually(func() string {
				// ctrlErrors.TestEmpty()
				RPFound = &v4alpha1.CloudflareAccessReusablePolicy{}
				_ = k8sClient.Get(ctx, RPTypeNamespaceName, RPFound)
				return RPFound.Status.AccessReusablePolicyID
			}).WithTimeout(defaultTimeout).WithPolling(defaultPoolRate).Should(Not(BeEmpty()))

			//
			//
			//

			By("Creating the custom resource for the Kind CloudflareAccessApplication")
			typeNamespaceName := types.NamespacedName{Name: "cloudflare-app-two", Namespace: cloudflareName}

			apps := &v4alpha1.CloudflareAccessApplication{
				ObjectMeta: metav1.ObjectMeta{
					Name:      typeNamespaceName.Name,
					Namespace: namespace.Name,
				},
				Spec: v4alpha1.CloudflareAccessApplicationSpec{
					Name:   "integration policies test",
					Domain: "integration-policies.cf-operator-tests.uk",
					PolicyRefs: []string{
						RPTypeNamespaceName.Name,
					},
				},
			}
			err = k8sClient.Create(ctx, apps)
			Expect(err).To(Not(HaveOccurred()))

			found := &v4alpha1.CloudflareAccessApplication{}
			By("Checking the latest Status should have the ID of the resource")
			Eventually(func() string {
				// ctrlErrors.TestEmpty()
				found = &v4alpha1.CloudflareAccessApplication{}
				_ = k8sClient.Get(ctx, typeNamespaceName, found)
				return found.Status.AccessApplicationID
			}).WithTimeout(defaultTimeout).WithPolling(defaultPoolRate).Should(Not(BeEmpty()))

			//
			//
			//

			By("Ensuring Cloudflare Application refers Reusable policy CF ID")
			Expect(found.Status.ReusablePolicyIDs).To(ContainElement(RPFound.Status.AccessReusablePolicyID))
		})

		It("should successfully reconcile and delete a custom resource for CloudflareAccessApplication", func() {
			By("Creating the custom resource for the Kind CloudflareAccessApplication")
			typeNamespaceName := types.NamespacedName{Name: "cloudflare-app-five", Namespace: cloudflareName}

			apps := &v4alpha1.CloudflareAccessApplication{
				ObjectMeta: metav1.ObjectMeta{
					Name:      typeNamespaceName.Name,
					Namespace: typeNamespaceName.Namespace,
				},
				Spec: v4alpha1.CloudflareAccessApplicationSpec{
					Name:   "integration test",
					Domain: "integration.cf-operator-tests.uk",
				},
			}
			err := k8sClient.Create(ctx, apps)
			Expect(err).To(Not(HaveOccurred()))

			By("Checking if the custom resource was successfully created")
			Eventually(func() error {
				// ctrlErrors.TestEmpty()
				found := &v4alpha1.CloudflareAccessApplication{}
				return k8sClient.Get(ctx, typeNamespaceName, found)
			}).WithTimeout(defaultTimeout).WithPolling(defaultPoolRate).Should(Succeed())

			found := &v4alpha1.CloudflareAccessApplication{}
			By("Checking the latest Status should have the ID of the resource")
			Eventually(func(g Gomega) { //nolint:varnamelen
				// ctrlErrors.TestEmpty()
				found = &v4alpha1.CloudflareAccessApplication{}
				g.Expect(k8sClient.Get(ctx, typeNamespaceName, found)).To(Not(HaveOccurred()))
				g.Expect(found.Status.AccessApplicationID).ToNot(BeEmpty())
				g.Expect(found.Status.CreatedAt.Time).To(Equal(found.Status.UpdatedAt.Time))
			}).WithTimeout(defaultTimeout).WithPolling(defaultPoolRate).Should(Succeed())

			By("Cloudflare resource should equal the spec")
			cfResource, err := api.AccessApplication(ctx, found.Status.AccessApplicationID)
			Expect(err).To(Not(HaveOccurred()))
			Expect(cfResource.Name).To(Equal(found.Spec.Name))

			By("Get the latest version of the resource")
			Expect(k8sClient.Get(ctx, typeNamespaceName, found)).To(Not(HaveOccurred()))
			By("Updating the name of the resource")

			found.Spec.Name = updtdName
			Expect(k8sClient.Update(ctx, found)).To(Not(HaveOccurred()))

			By("Checking the latest Status should have the update")
			Eventually(func(g Gomega) { //nolint:varnamelen
				// ctrlErrors.TestEmpty()
				found = &v4alpha1.CloudflareAccessApplication{}
				g.Expect(k8sClient.Get(ctx, typeNamespaceName, found)).To(Not(HaveOccurred()))
				g.Expect(found.Spec.Name).To(Equal(updtdName))
				g.Expect(found.Status.AccessApplicationID).ToNot(BeEmpty())
			}).WithTimeout(defaultTimeout).WithPolling(defaultPoolRate).Should(Succeed())

			By("Cloudflare resource should equal the updated spec")
			Eventually(func(g Gomega) { //nolint:varnamelen
				// ctrlErrors.TestEmpty()
				cfResource, err = api.AccessApplication(ctx, found.Status.AccessApplicationID)
				g.Expect(err).To(Not(HaveOccurred()))
				g.Expect(cfResource.Name).To(Equal(found.Spec.Name))
			}).WithTimeout(defaultTimeout).WithPolling(defaultPoolRate).Should(Succeed(), ctrlErrors) // sometimes this is cached

			By("Cloudflare resource should be deleted")
			Expect(k8sClient.Delete(ctx, apps)).To(Not(HaveOccurred()))

			By("Checking if the custom resource was successfully deleted")
			Eventually(func() error {
				// ctrlErrors.TestEmpty()
				return k8sClient.Get(ctx, typeNamespaceName, apps)
			}).WithTimeout(defaultTimeout).WithPolling(defaultPoolRate).Should(Not(Succeed()))
		})

		It("should be able to set a LogoURL for CloudflareAccessApplication", func() {
			By("Creating the custom resource for the Kind CloudflareAccessApplication")
			typeNamespaceName := types.NamespacedName{Name: "cloudflare-app-six", Namespace: cloudflareName}

			apps := &v4alpha1.CloudflareAccessApplication{
				ObjectMeta: metav1.ObjectMeta{
					Name:      typeNamespaceName.Name,
					Namespace: typeNamespaceName.Namespace,
				},
				Spec: v4alpha1.CloudflareAccessApplicationSpec{
					Name:    "integration test logo",
					Domain:  "integration-logo-test.cf-operator-tests.uk",
					LogoURL: "https://www.cloudflare.com/img/logo-web-badges/cf-logo-on-white-bg.svg",
				},
			}
			err := k8sClient.Create(ctx, apps)
			Expect(err).To(Not(HaveOccurred()))

			By("Checking if the custom resource was successfully created")
			Eventually(func() error {
				// ctrlErrors.TestEmpty()
				found := &v4alpha1.CloudflareAccessApplication{}
				return k8sClient.Get(ctx, typeNamespaceName, found)
			}).WithTimeout(defaultTimeout).WithPolling(defaultPoolRate).Should(Succeed())

			found := &v4alpha1.CloudflareAccessApplication{}
			By("Checking the latest Status should have the ID of the resource")
			Eventually(func(g Gomega) { //nolint:varnamelen
				// ctrlErrors.TestEmpty()
				found = &v4alpha1.CloudflareAccessApplication{}
				g.Expect(k8sClient.Get(ctx, typeNamespaceName, found)).To(Not(HaveOccurred()))
				g.Expect(found.Status.AccessApplicationID).ToNot(BeEmpty())
				g.Expect(found.Status.CreatedAt.Time).To(Equal(found.Status.UpdatedAt.Time))
			}).WithTimeout(defaultTimeout).WithPolling(defaultPoolRate).Should(Succeed())

			By("Cloudflare resource should equal the spec")
			cfResource, err := api.AccessApplication(ctx, found.Status.AccessApplicationID)
			Expect(err).To(Not(HaveOccurred()))
			Expect(cfResource.Name).To(Equal(found.Spec.Name))
		})

		It("should successfully reconcile CloudflareAccessApplication whose AccessApplicationID references a missing Application", func() {
			By("Recreating the custom resource for the Kind CloudflareAccessApplication")
			typeNamespaceName := types.NamespacedName{Name: "cloudflare-app-seven", Namespace: cloudflareName}

			previousCreatedAndUpdatedDate := metav1.NewTime(time.Now().Add(-time.Hour * 24))
			apps := &v4alpha1.CloudflareAccessApplication{
				ObjectMeta: metav1.ObjectMeta{
					Name:      typeNamespaceName.Name,
					Namespace: namespace.Name,
				},
				Spec: v4alpha1.CloudflareAccessApplicationSpec{
					Name:   "missing application",
					Domain: "recreate-application.cf-operator-tests.uk",
				},
			}

			err := k8sClient.Create(ctx, apps)
			Expect(err).To(Not(HaveOccurred()))

			By("Checking if the custom resource was successfully created")
			Eventually(func() error {
				// ctrlErrors.TestEmpty()
				found := &v4alpha1.CloudflareAccessApplication{}
				return k8sClient.Get(ctx, typeNamespaceName, found)
			}).WithTimeout(defaultTimeout).WithPolling(defaultPoolRate).Should(Succeed())

			found := &v4alpha1.CloudflareAccessApplication{}
			By("Checking the latest Status should have the ID of the resource")

			oldAccessApplicationID := ""

			Eventually(func(g Gomega) { //nolint:varnamelen
				// ctrlErrors.TestEmpty()
				found = &v4alpha1.CloudflareAccessApplication{}
				g.Expect(k8sClient.Get(ctx, typeNamespaceName, found)).To(Not(HaveOccurred()))
				g.Expect(found.Status.AccessApplicationID).ToNot(BeEmpty())
				oldAccessApplicationID = found.Status.AccessApplicationID
				g.Expect(found.Status.CreatedAt.Time).To(Equal(found.Status.UpdatedAt.Time))
				g.Expect(found.Status.CreatedAt.Time.After(previousCreatedAndUpdatedDate.Time)).To(BeTrue())
				g.Expect(found.Status.UpdatedAt.Time.After(previousCreatedAndUpdatedDate.Time)).To(BeTrue())
			}).WithTimeout(defaultTimeout).WithPolling(defaultPoolRate).Should(Succeed())

			Expect(api.DeleteAccessApplication(ctx, found.Status.AccessApplicationID)).To(Not(HaveOccurred()))

			By("re-trigger reconcile by updating access application")
			Eventually(func(g Gomega) { //nolint:varnamelen
				// ctrlErrors.TestEmpty()
				found = &v4alpha1.CloudflareAccessApplication{}
				g.Expect(k8sClient.Get(ctx, typeNamespaceName, found)).To(Not(HaveOccurred()))
				found.Spec.Name = updtdName
				Expect(k8sClient.Update(ctx, found)).To(Not(HaveOccurred()))
			}).WithTimeout(defaultTimeout).WithPolling(defaultPoolRate).Should(Succeed())

			By("Checking the latest Status should have the ID of the resource")
			Eventually(func(g Gomega) { //nolint:varnamelen
				// ctrlErrors.TestEmpty()
				found = &v4alpha1.CloudflareAccessApplication{}
				g.Expect(k8sClient.Get(ctx, typeNamespaceName, found)).To(Not(HaveOccurred()))
				g.Expect(found.Status.AccessApplicationID).ToNot(Equal(oldAccessApplicationID))
			}).WithTimeout(defaultTimeout).WithPolling(defaultPoolRate).Should(Succeed())

			By("Cloudflare resource should equal the updated spec")
			Eventually(func(g Gomega) { //nolint:varnamelen
				// ctrlErrors.TestEmpty()
				cfResource, err := api.AccessApplication(ctx, found.Status.AccessApplicationID)
				g.Expect(err).To(Not(HaveOccurred()))
				g.Expect(cfResource.Name).To(Equal(found.Spec.Name))
			}).WithTimeout(defaultTimeout).WithPolling(defaultPoolRate).Should(Succeed(), ctrlErrors) // sometimes this is cached
		})
	})
})
