// TODO: add back //go:build integration

package controller_test

import (
	"context"

	"github.com/bojanzelic/cloudflare-zero-trust-operator/api/v4alpha1"

	"github.com/bojanzelic/cloudflare-zero-trust-operator/internal/cftypes"
	"github.com/bojanzelic/cloudflare-zero-trust-operator/internal/meta"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("CloudflareServiceToken controller", Ordered, func() {
	BeforeAll(func() { insertedTracer.ResetCFUUIDs() })
	AfterAll(func() { insertedTracer.UninstallFromCF(api) })

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
			// Expect(err).To(Not(HaveOccurred()))
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

		It("should successfully reconcile a custom resource for CloudflareServiceToken", func() {
			By("Creating the custom resource for the Kind CloudflareServiceToken")
			sTokenNN := types.NamespacedName{Name: "test-1-stoken", Namespace: testScopedNamespace}
			serviceToken := &v4alpha1.CloudflareServiceToken{
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

			Expect(k8sClient.Create(ctx, serviceToken)).To(Not(HaveOccurred()))

			By("Checking if the secret was successfully created")
			sec := &corev1.Secret{}
			expectedSecondarySecIdentity := types.NamespacedName{Name: serviceToken.Spec.Template.Name, Namespace: serviceToken.Namespace}
			Eventually(func() error {
				// ctrlErrors.TestEmpty()
				return k8sClient.Get(ctx, expectedSecondarySecIdentity, sec)
			}).WithTimeout(defaultTimeout).WithPolling(defaultPollRate).Should(Succeed())

			By("Make sure the status ref is what we expect")
			Eventually(func(g Gomega) { //nolint:varnamelen
				// ctrlErrors.TestEmpty()
				err := k8sClient.Get(ctx, sTokenNN, serviceToken)
				g.Expect(err).ToNot(HaveOccurred())

				//
				g.Expect(serviceToken.Status.ServiceTokenID).ToNot(BeEmpty())
				g.Expect(serviceToken.Status.SecretRef).ToNot(BeNil())
				g.Expect(serviceToken.Status.SecretRef.ClientIDKey).ToNot(BeEmpty())
				g.Expect(serviceToken.Status.SecretRef.ClientSecretKey).ToNot(BeEmpty())
				g.Expect(serviceToken.Status.SecretRef.Name).To(Equal(sec.Name))
			}).WithTimeout(defaultTimeout).WithPolling(defaultPollRate).Should(Succeed())

			expectedID := string(sec.Data[sec.Annotations[meta.AnnotationTokenIDKey]])
			expectedClientID := string(sec.Data[sec.Annotations[meta.AnnotationClientIDKey]])

			By("Checking if the resource exists in cloudflare")
			cfst, err := api.AccessServiceToken(ctx, expectedID)
			Expect(err).To(Not(HaveOccurred()))
			Expect(cfst.ID).To(Equal(expectedID))
			Expect(cfst.ClientID).To(Equal(expectedClientID))

			By("Renaming the secret")
			serviceToken.Spec.Name = "updated_secret_name"
			err = k8sClient.Update(ctx, serviceToken)
			Expect(err).ToNot(HaveOccurred())

			By("Awaiting update instruction being acknoledged")
			Eventually(func(g Gomega) { //nolint:varnamelen
				// ctrlErrors.TestEmpty()
				tmp := &v4alpha1.CloudflareServiceToken{}
				err = k8sClient.Get(ctx, sTokenNN, tmp)
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(tmp.Status.UpdatedAt.Time).To(BeTemporally(">", serviceToken.Status.UpdatedAt.Time))
			}).WithTimeout(defaultTimeout).WithPolling(defaultPollRate).Should(Succeed())

			//
			By("Expecting changes NOT to happen in CF Service Token name (Updates not implemented)")
			cfst, err = api.AccessServiceToken(ctx, expectedID)
			Expect(err).To(Not(HaveOccurred()))
			Expect(cfst.Name).To(Equal(serviceToken.Spec.Name))
		})

		It("should successfully reconcile a custom resource for CloudflareServiceToken", func() {
			By("Creating the custom resource for the Kind CloudflareServiceToken")
			var sToken *v4alpha1.CloudflareServiceToken
			sTokenNN := types.NamespacedName{Name: "test-2-stoken", Namespace: testScopedNamespace}
			token := &v4alpha1.CloudflareServiceToken{}
			err := k8sClient.Get(ctx, sTokenNN, token)
			if err != nil && errors.IsNotFound(err) {
				sToken = &v4alpha1.CloudflareServiceToken{
					ObjectMeta: metav1.ObjectMeta{
						Name:      sTokenNN.Name,
						Namespace: sTokenNN.Namespace,
					},
					Spec: v4alpha1.CloudflareServiceTokenSpec{
						Name: "ZTO AccessServiceToken Tests - 2 - SToken",
					},
				}
				Expect(k8sClient.Create(ctx, sToken)).To(Not(HaveOccurred()))
			}

			By("Checking if the custom resource was successfully created")
			found := &v4alpha1.CloudflareServiceToken{}
			Eventually(func() error {
				// ctrlErrors.TestEmpty()
				return k8sClient.Get(ctx, sTokenNN, found)
			}).WithTimeout(defaultTimeout).WithPolling(defaultPollRate).Should(Succeed())

			Expect(err).To(Not(HaveOccurred()))

			By("Checking to get the updated CR")
			Eventually(func() error {
				// ctrlErrors.TestEmpty()
				return k8sClient.Get(ctx, sTokenNN, found)
			}).WithTimeout(defaultTimeout).WithPolling(defaultPollRate).Should(Succeed())

			By("Make sure the status ref is what we expect")
			Eventually(func(g Gomega) { //nolint:varnamelen
				// ctrlErrors.TestEmpty()
				_ = k8sClient.Get(ctx, sTokenNN, found)
				g.Expect(found.Status.ServiceTokenID).ToNot(BeEmpty())
				g.Expect(found.Status.SecretRef).ToNot(BeNil())
				g.Expect(found.Status.SecretRef.ClientIDKey).ToNot(BeEmpty())
				g.Expect(found.Status.SecretRef.ClientSecretKey).ToNot(BeEmpty())
				g.Expect(found.Status.SecretRef.Name).ToNot(BeEmpty())
			}).WithTimeout(defaultTimeout).WithPolling(defaultPollRate).Should(Succeed())

			sec := &corev1.Secret{}
			By("Making sure that the secret exists")
			Eventually(func() error {
				// ctrlErrors.TestEmpty()
				return k8sClient.Get(ctx, sTokenNN, sec)
			}).WithTimeout(defaultTimeout).WithPolling(defaultPollRate).Should(Succeed())

			By("Checking if the resource exists in cloudflare")
			tokens, err := api.AccessServiceTokens(ctx)
			Expect(err).To(Not(HaveOccurred()))

			secretFound := false
			for _, token := range *tokens {
				if token.Name == found.Spec.Name {
					secretFound = true

					Expect(token.ID).To(Equal(string(sec.Data[sec.Annotations[meta.AnnotationTokenIDKey]])))
					Expect(token.ClientID).To(Equal(string(sec.Data[sec.Annotations[meta.AnnotationClientIDKey]])))
				}
			}
			// we should only have 1 Token created

			Expect(secretFound).To(BeTrue(), "secret not found", found.Spec.Name, tokens)

			By("Updating the service token to move the secret")
			_ = k8sClient.Get(ctx, sTokenNN, found)
			found.Spec.Template.Name = "moved-secret"
			Eventually(func() error {
				// ctrlErrors.TestEmpty()
				return k8sClient.Update(ctx, found)
			}).WithTimeout(defaultTimeout).WithPolling(defaultPollRate).Should(Succeed())

			By("Checking if the new secret was successfully created")
			Eventually(func() error {
				// ctrlErrors.TestEmpty()
				sec = &corev1.Secret{}
				return k8sClient.Get(ctx, types.NamespacedName{Name: found.Spec.Template.Name, Namespace: sTokenNN.Namespace}, sec)
			}).WithTimeout(defaultTimeout).WithPolling(defaultPollRate).Should(Succeed())

			By("Make sure the status ref is what we expect")
			_ = k8sClient.Get(ctx, sTokenNN, found)
			Expect(found.Status.ServiceTokenID).ToNot(BeEmpty())
			Expect(found.Status.SecretRef).ToNot(BeNil())
			Expect(found.Status.SecretRef.ClientIDKey).ToNot(BeEmpty())
			Expect(found.Status.SecretRef.ClientSecretKey).ToNot(BeEmpty())
			Expect(found.Status.SecretRef.Name).ToNot(BeEmpty())

			By("Updating the secret template")
			err = k8sClient.Get(ctx, sTokenNN, sToken)
			Expect(err).ToNot(HaveOccurred())
			sToken.Spec.Template.Name = "moved-secret"
			sToken.Spec.Template.ClientIDKey = "keylocation"
			err = k8sClient.Update(ctx, sToken)
			Expect(err).ToNot(HaveOccurred())

			By("Make sure the status ref is what we expect")
			Eventually(func(g Gomega) { //nolint:varnamelen
				// ctrlErrors.TestEmpty()
				err = k8sClient.Get(ctx, sTokenNN, sToken)
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(sToken.Status.SecretRef.Name).To(Equal(sToken.Spec.Template.Name))
				g.Expect(sToken.Status.SecretRef.ClientIDKey).To(Equal(sToken.Spec.Template.ClientIDKey))
			}).Should(Succeed())

			By("Checking if the new secret was successfully created")
			Eventually(func() error {
				// ctrlErrors.TestEmpty()
				sec = &corev1.Secret{}
				return k8sClient.Get(ctx, types.NamespacedName{Name: sToken.Spec.Template.Name, Namespace: sTokenNN.Namespace}, sec)
			}).WithTimeout(defaultTimeout).WithPolling(defaultPollRate).Should(Succeed())
			Expect(sec.Data).To(HaveKey(sToken.Spec.Template.ClientIDKey))

			By("Checking if the old secret was removed")
			Eventually(func() error {
				// ctrlErrors.TestEmpty()
				sec := &corev1.Secret{}
				return k8sClient.Get(ctx, sTokenNN, sec)
			}).WithTimeout(defaultTimeout).WithPolling(defaultPollRate).ShouldNot(Succeed())
		})

		It("should successfully allow removal of resource if externally deleted", func() {
			By("Creating the custom resource for the Kind CloudflareServiceToken")
			token := &v4alpha1.CloudflareServiceToken{}
			sTokenNN := types.NamespacedName{Name: "test-3-stoken", Namespace: testScopedNamespace}
			err := k8sClient.Get(ctx, sTokenNN, token)
			if err != nil && errors.IsNotFound(err) {
				token = &v4alpha1.CloudflareServiceToken{
					ObjectMeta: metav1.ObjectMeta{
						Name:      sTokenNN.Name,
						Namespace: sTokenNN.Namespace,
					},
					Spec: v4alpha1.CloudflareServiceTokenSpec{
						Name: "ZTO AccessServiceToken Tests - 3 - SToken",
					},
				}
				Expect(k8sClient.Create(ctx, token)).To(Not(HaveOccurred()))
			}

			By("Checking to get the updated CR")
			Eventually(func() error {
				// ctrlErrors.TestEmpty()
				return k8sClient.Get(ctx, sTokenNN, token)
			}).WithTimeout(defaultTimeout).WithPolling(defaultPollRate).Should(Succeed())

			By("Make sure the status ref is what we expect")
			Eventually(func(g Gomega) { //nolint:varnamelen
				// ctrlErrors.TestEmpty()
				_ = k8sClient.Get(ctx, sTokenNN, token)
				g.Expect(token.Status.ServiceTokenID).ToNot(BeEmpty())
			}).WithTimeout(defaultTimeout).WithPolling(defaultPollRate).Should(Succeed())

			By("Externally removing the token")
			Expect(api.DeleteAccessServiceToken(ctx, token.Status.ServiceTokenID)).To(Succeed())

			By("Removing the access service token")
			Expect(k8sClient.Delete(ctx, token)).To(Succeed())
		})

		It("should successfully reconcile a custom resource for CloudflareServiceToken", func() {
			By("Creating the custom resource for the Kind CloudflareServiceToken")
			token := &v4alpha1.CloudflareServiceToken{}
			sTokenNN := types.NamespacedName{Name: "test-4-stoken", Namespace: testScopedNamespace}
			err := k8sClient.Get(ctx, sTokenNN, token)
			if err != nil && errors.IsNotFound(err) {
				token = &v4alpha1.CloudflareServiceToken{
					ObjectMeta: metav1.ObjectMeta{
						Name:      sTokenNN.Name,
						Namespace: sTokenNN.Namespace,
					},
					Spec: v4alpha1.CloudflareServiceTokenSpec{
						Name: "ZTO AccessServiceToken Tests - 4 - SToken",
					},
				}
				Expect(k8sClient.Create(ctx, token)).To(Not(HaveOccurred()))
			}

			By("Checking to get the updated CR")
			Eventually(func() error {
				// ctrlErrors.TestEmpty()
				return k8sClient.Get(ctx, sTokenNN, token)
			}).WithTimeout(defaultTimeout).WithPolling(defaultPollRate).Should(Succeed())

			By("Make sure the annotation is not present")
			Eventually(func(g Gomega) { //nolint:varnamelen
				// ctrlErrors.TestEmpty()
				_ = k8sClient.Get(ctx, sTokenNN, token)
				g.Expect(token.Status.ServiceTokenID).ToNot(BeEmpty())
				keyValue, keyExists := token.Annotations[meta.AnnotationPreventDestroy]
				g.Expect(keyExists).To(BeFalse())
				g.Expect(keyValue).To(Or(BeEmpty(), Equal("false")))
			}).WithTimeout(defaultTimeout).WithPolling(defaultPollRate).Should(Succeed())

			By("Make sure the service token exists on cloudflare")
			tokens, err := api.AccessServiceTokens(ctx)
			Expect(err).To(Not(HaveOccurred()))
			var foundToken *cftypes.ExtendedServiceToken
			for _, cfToken := range *tokens {
				if cfToken.ID == token.Status.ServiceTokenID {
					foundToken = &cfToken
				}
			}

			Expect(foundToken).ToNot(BeNil())

			By("Removing the access service token")
			_ = k8sClient.Delete(ctx, token)

			By("Expecting that the token is removed from cloudflare")
			Eventually(func(g Gomega) { //nolint:varnamelen
				// ctrlErrors.TestEmpty()
				tokens, _ := api.AccessServiceTokens(ctx)
				var foundToken *cftypes.ExtendedServiceToken
				for _, cfToken := range *tokens {
					if cfToken.ID == token.Status.ServiceTokenID {
						foundToken = &cfToken
					}
				}

				g.Expect(foundToken).To(BeNil())
			}).WithTimeout(defaultTimeout).WithPolling(defaultPollRate).Should(Succeed())
		})

		It("should successfully not remove the resource in CF if annotation is set", func() {
			By("Creating the custom resource for the Kind CloudflareServiceToken")
			token := &v4alpha1.CloudflareServiceToken{}
			sTokenNN := types.NamespacedName{Name: "test-5-stoken", Namespace: testScopedNamespace}
			err := k8sClient.Get(ctx, sTokenNN, token)
			if err != nil && errors.IsNotFound(err) {
				token = &v4alpha1.CloudflareServiceToken{
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
				Expect(k8sClient.Create(ctx, token)).To(Not(HaveOccurred()))
			}

			By("Checking to get the updated CR")
			Eventually(func() error {
				// ctrlErrors.TestEmpty()
				return k8sClient.Get(ctx, sTokenNN, token)
			}).WithTimeout(defaultTimeout).WithPolling(defaultPollRate).Should(Succeed())

			By("Make sure the status ref is what we expect")
			Eventually(func(g Gomega) { //nolint:varnamelen
				// ctrlErrors.TestEmpty()
				_ = k8sClient.Get(ctx, sTokenNN, token)
				g.Expect(token.Status.ServiceTokenID).ToNot(BeEmpty())
			}).WithTimeout(defaultTimeout).WithPolling(defaultPollRate).Should(Succeed())

			By("Make sure the service token exists on cloudflare")
			tokens, err := api.AccessServiceTokens(ctx)
			Expect(err).To(Not(HaveOccurred()))
			var foundToken *cftypes.ExtendedServiceToken
			for _, cfToken := range *tokens {
				if cfToken.ID == token.Status.ServiceTokenID {
					foundToken = &cfToken
				}
			}

			Expect(foundToken).ToNot(BeNil())

			By("Removing the access service token")
			_ = k8sClient.Delete(ctx, token)

			By("Make sure the service token exists on cloudflare")
			tokens, err = api.AccessServiceTokens(ctx)
			Expect(err).To(Not(HaveOccurred()))
			foundToken = nil
			for _, cfToken := range *tokens {
				if cfToken.ID == token.Status.ServiceTokenID {
					foundToken = &cfToken
				}
			}

			Expect(foundToken).ToNot(BeNil())
		})

	})
})
