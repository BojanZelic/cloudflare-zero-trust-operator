// TODO: add back //go:build integration

package controller_test

import (
	"context"

	"github.com/bojanzelic/cloudflare-zero-trust-operator/api/v4alpha1"

	"github.com/bojanzelic/cloudflare-zero-trust-operator/internal/meta"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("CloudflareServiceToken controller", Ordered, func() {
	BeforeAll(func() { insertedTracer.ResetStores() })
	AfterAll(func() {
		errs := insertedTracer.UninstallFromCF(api)
		Expect(errs).To(BeEmpty())
	})

	//
	//
	//

	Context("CloudflareServiceToken controller test", func() {
		const testScopedNamespace = "zto-testing-stoken"

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
		//
		//

		It("should create and validate a CloudflareServiceToken custom resource with secret creation and renaming", func() {
			By("Creating the custom resource for the Kind CloudflareServiceToken")
			sTokenNN := types.NamespacedName{Name: "test-1-stoken", Namespace: testScopedNamespace}
			sToken := &v4alpha1.CloudflareServiceToken{
				ObjectMeta: metav1.ObjectMeta{
					Name:      sTokenNN.Name,
					Namespace: sTokenNN.Namespace,
				},
				Spec: v4alpha1.CloudflareServiceTokenSpec{
					Name: "ZTO AccessServiceToken Tests - 1 - SToken",
					Template: v4alpha1.SecretTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Name: "secret-location",
						},
					},
				},
			}
			Expect(k8sClient.Create(ctx, sToken)).ToNot(HaveOccurred())

			//
			ByExpectingCFResourceToBeReady(ctx, sToken).Should(Succeed())

			By("Checking if the secret was successfully created")
			sTokenSecret := &corev1.Secret{}
			Eventually(func() error {
				// ctrlErrors.TestEmpty()
				return k8sClient.Get(ctx, sToken.GetSecretNamespacedName(), sTokenSecret)
			}).WithTimeout(defaultTimeout).WithPolling(defaultPollRate).Should(Succeed())

			By("Make sure the status ref is what we expect")
			Eventually(func(g Gomega) { //nolint:varnamelen
				// ctrlErrors.TestEmpty()
				err := k8sClient.Get(ctx, sTokenNN, sToken)
				g.Expect(err).ToNot(HaveOccurred())

				//
				g.Expect(sToken.GetCloudflareUUID()).ToNot(BeEmpty())
				g.Expect(sToken.Status.SecretRef).ToNot(BeNil())
				g.Expect(sToken.Status.SecretRef.ClientIDKey).ToNot(BeEmpty())
				g.Expect(sToken.Status.SecretRef.ClientSecretKey).ToNot(BeEmpty())
				g.Expect(sToken.Status.SecretRef.Name).To(Equal(sTokenSecret.Name))
			}).WithTimeout(defaultTimeout).WithPolling(defaultPollRate).Should(Succeed())

			expectedID := string(sTokenSecret.Data[sTokenSecret.Annotations[meta.AnnotationTokenIDKey]])
			expectedClientID := string(sTokenSecret.Data[sTokenSecret.Annotations[meta.AnnotationClientIDKey]])

			By("Checking if the resource exists in cloudflare")
			cfst, err := api.AccessServiceToken(ctx, expectedID)
			Expect(err).ToNot(HaveOccurred())
			Expect(cfst.ID).To(Equal(expectedID))
			Expect(cfst.ClientID).To(Equal(expectedClientID))

			By("Renaming the service token name")
			addDirtyingSuffix(&sToken.Spec.Name)
			Expect(k8sClient.Update(ctx, sToken)).ToNot(HaveOccurred())

			// Await for resource to be ready again
			ByExpectingCFResourceToBeReady(ctx, sToken).Should(Succeed())

			//
			By("Expecting name NOT to change on CloudFlare's side - (Updates on Service Token not implemented)")
			cfst, err = api.AccessServiceToken(ctx, expectedID)
			Expect(err).ToNot(HaveOccurred())
			Expect(cfst.Name).ToNot(Equal(sToken.Spec.Name))
		})

		It("should create a CloudflareServiceToken, verify secret creation, and test secret relocation and key updates", func() {
			By("Creating the custom resource for the Kind CloudflareServiceToken")
			sTokenNN := types.NamespacedName{Name: "test-2-stoken", Namespace: testScopedNamespace}
			sToken := &v4alpha1.CloudflareServiceToken{
				ObjectMeta: metav1.ObjectMeta{
					Name:      sTokenNN.Name,
					Namespace: sTokenNN.Namespace,
				},
				Spec: v4alpha1.CloudflareServiceTokenSpec{
					Name: "ZTO AccessServiceToken Tests - 2 - SToken",
				},
			}
			Expect(k8sClient.Create(ctx, sToken)).ToNot(HaveOccurred())

			//
			ByExpectingCFResourceToBeReady(ctx, sToken).Should(Succeed())

			Expect(sToken.GetCloudflareUUID()).ToNot(BeEmpty())
			Expect(sToken.Status.SecretRef).ToNot(BeNil())
			Expect(sToken.Status.SecretRef.ClientIDKey).ToNot(BeEmpty())
			Expect(sToken.Status.SecretRef.ClientSecretKey).ToNot(BeEmpty())
			Expect(sToken.Status.SecretRef.Name).ToNot(BeEmpty())

			By("Making sure that the secret exists")
			sTokenSecret := &corev1.Secret{}
			Eventually(func() error {
				// ctrlErrors.TestEmpty()
				return k8sClient.Get(ctx, sTokenNN, sTokenSecret)
			}).WithTimeout(defaultTimeout).WithPolling(defaultPollRate).Should(Succeed())

			By("Checking if the resource exists in cloudflare")
			cfSToken, err := api.AccessServiceToken(ctx, sToken.GetCloudflareUUID())
			Expect(err).ToNot(HaveOccurred())
			Expect(cfSToken.Name).To(Equal(sToken.Spec.Name))
			Expect(cfSToken.ID).To(Equal(string(sTokenSecret.Data[sTokenSecret.Annotations[meta.AnnotationTokenIDKey]])))
			Expect(cfSToken.ClientID).To(Equal(string(sTokenSecret.Data[sTokenSecret.Annotations[meta.AnnotationClientIDKey]])))

			By("Updating the service token to move the secret")
			sToken.Spec.Template.Name = "moved-secret"
			Expect(k8sClient.Update(ctx, sToken)).ToNot(HaveOccurred())

			// Await for resource to be ready again
			ByExpectingCFResourceToBeReady(ctx, sToken).Should(Succeed())

			By("Checking if the new secret was successfully created")
			sTokenSecret = &corev1.Secret{}
			Eventually(func() error {
				// ctrlErrors.TestEmpty()
				return k8sClient.Get(ctx, sToken.GetSecretNamespacedName(), sTokenSecret)
			}).WithTimeout(defaultTimeout).WithPolling(defaultPollRate).Should(Succeed())

			By("Make sure the status ref is what we expect")
			_ = k8sClient.Get(ctx, sTokenNN, sToken)
			Expect(sToken.GetCloudflareUUID()).ToNot(BeEmpty())
			Expect(sToken.Status.SecretRef).ToNot(BeNil())
			Expect(sToken.Status.SecretRef.ClientIDKey).ToNot(BeEmpty())
			Expect(sToken.Status.SecretRef.ClientSecretKey).ToNot(BeEmpty())
			Expect(sToken.Status.SecretRef.Name).ToNot(BeEmpty())

			By("Updating the secret template")
			sToken.Spec.Template.Name = "moved-secret"
			sToken.Spec.Template.ClientIDKey = "keylocation"
			Expect(k8sClient.Update(ctx, sToken)).ToNot(HaveOccurred())

			// Await for resource to be ready again
			ByExpectingCFResourceToBeReady(ctx, sToken).Should(Succeed())

			By("Make sure the status ref is what we expect")
			Expect(sToken.Status.SecretRef.Name).To(Equal(sToken.Spec.Template.Name))
			Expect(sToken.Status.SecretRef.ClientIDKey).To(Equal(sToken.Spec.Template.ClientIDKey))

			By("Checking if the new secret was successfully created")
			sTokenSecret = &corev1.Secret{}
			Eventually(func() error {
				// ctrlErrors.TestEmpty()

				return k8sClient.Get(ctx, sToken.GetSecretNamespacedName(), sTokenSecret)
			}).WithTimeout(defaultTimeout).WithPolling(defaultPollRate).Should(Succeed())
			Expect(sTokenSecret.Data).To(HaveKey(sToken.Spec.Template.ClientIDKey))

			By("Checking if the old secret was removed")
			sTokenSecret = &corev1.Secret{}
			Eventually(func() error {
				// ctrlErrors.TestEmpty()

				return k8sClient.Get(ctx, sTokenNN, sTokenSecret)
			}).WithTimeout(defaultTimeout).WithPolling(defaultPollRate).ShouldNot(Succeed())
		})

		It("should create, verify, and delete a CloudflareServiceToken custom resource, ensuring Cloudflare cleanup", func() {
			By("Creating the custom resource for the Kind CloudflareServiceToken")
			sTokenNN := types.NamespacedName{Name: "test-3-stoken", Namespace: testScopedNamespace}
			sToken := &v4alpha1.CloudflareServiceToken{
				ObjectMeta: metav1.ObjectMeta{
					Name:      sTokenNN.Name,
					Namespace: sTokenNN.Namespace,
				},
				Spec: v4alpha1.CloudflareServiceTokenSpec{
					Name: "ZTO AccessServiceToken Tests - 3 - SToken",
				},
			}
			Expect(k8sClient.Create(ctx, sToken)).ToNot(HaveOccurred())

			//
			ByExpectingCFResourceToBeReady(ctx, sToken).Should(Succeed())

			By("Make sure the annotation is not present")
			Eventually(func(g Gomega) { //nolint:varnamelen
				// ctrlErrors.TestEmpty()
				_ = k8sClient.Get(ctx, sTokenNN, sToken)
				g.Expect(sToken.GetCloudflareUUID()).ToNot(BeEmpty())
				keyValue, keyExists := sToken.Annotations[meta.AnnotationPreventDestroy]
				g.Expect(keyExists).To(BeFalse())
				g.Expect(keyValue).To(Or(BeEmpty(), Equal("false")))
			}).WithTimeout(defaultTimeout).WithPolling(defaultPollRate).Should(Succeed())

			By("Make sure the service token exists on cloudflare")
			foundToken, err := api.AccessServiceToken(ctx, sToken.GetCloudflareUUID())
			Expect(err).ToNot(HaveOccurred())
			Expect(foundToken).ToNot(BeNil())

			By("Removing the access service token")
			Expect(k8sClient.Delete(ctx, sToken)).ToNot(HaveOccurred())

			//
			ByExpectingDeletionOf(sToken).Should(Succeed())

			By("Expecting that the token is removed from cloudflare")
			_, err = api.AccessServiceToken(ctx, sToken.GetCloudflareUUID())
			Expect(err).To(HaveOccurred())
			Expect(api.Is404(err)).To(BeTrue())
		})

		It("should successfully allow removal of resource if externally deleted", func() {
			By("Creating the custom resource for the Kind CloudflareServiceToken")
			sTokenNN := types.NamespacedName{Name: "test-4-stoken", Namespace: testScopedNamespace}
			sToken := &v4alpha1.CloudflareServiceToken{
				ObjectMeta: metav1.ObjectMeta{
					Name:      sTokenNN.Name,
					Namespace: sTokenNN.Namespace,
				},
				Spec: v4alpha1.CloudflareServiceTokenSpec{
					Name: "ZTO AccessServiceToken Tests - 4 - SToken",
				},
			}
			Expect(k8sClient.Create(ctx, sToken)).ToNot(HaveOccurred())

			//
			ByExpectingCFResourceToBeReady(ctx, sToken).Should(Succeed())

			By("Externally removing the token")
			Expect(api.DeleteAccessServiceToken(ctx, sToken.GetCloudflareUUID())).To(Succeed())

			By("Removing the access service token")
			Expect(k8sClient.Delete(ctx, sToken)).To(Succeed())

			//
			ByExpectingDeletionOf(sToken).Should(Succeed())
		})

		It("should successfully not remove the resource in CF if annotation is set", func() {
			By("Creating the custom resource for the Kind CloudflareServiceToken")
			sTokenNN := types.NamespacedName{Name: "test-5-stoken", Namespace: testScopedNamespace}
			sToken := &v4alpha1.CloudflareServiceToken{
				ObjectMeta: metav1.ObjectMeta{
					Name:      sTokenNN.Name,
					Namespace: sTokenNN.Namespace,
					Annotations: map[string]string{
						meta.AnnotationPreventDestroy: "true",
					},
				},
				Spec: v4alpha1.CloudflareServiceTokenSpec{
					Name: "ZTO AccessServiceToken Tests - 5 - SToken",
				},
			}
			Expect(k8sClient.Create(ctx, sToken)).ToNot(HaveOccurred())

			//
			ByExpectingCFResourceToBeReady(ctx, sToken).Should(Succeed())

			By("Make sure the service token exists on cloudflare")
			foundToken, err := api.AccessServiceToken(ctx, sToken.GetCloudflareUUID())
			Expect(err).ToNot(HaveOccurred())
			Expect(foundToken).ToNot(BeNil())

			By("Removing the access service token")
			Expect(k8sClient.Delete(ctx, sToken)).ToNot(HaveOccurred())

			//
			ByExpectingDeletionOf(sToken).Should(Succeed())

			By("Ensure service token still exists on CloudFlare")
			foundToken, err = api.AccessServiceToken(ctx, sToken.GetCloudflareUUID())
			Expect(err).ToNot(HaveOccurred())
			Expect(foundToken).ToNot(BeNil())
		})
	})
})
