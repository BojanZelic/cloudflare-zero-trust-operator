package controller

import (
	"context"
	"time"

	"github.com/bojanzelic/cloudflare-zero-trust-operator/api/v4alpha1"

	"github.com/bojanzelic/cloudflare-zero-trust-operator/internal/cftypes"
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

		const nsName = "servicetoken"

		ctx := context.Background()

		namespace := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name:      nsName,
				Namespace: nsName,
			},
		}

		BeforeEach(func() {
			logOutput.Clear()

			By("Creating the Namespace to perform the tests")
			_ = k8sClient.Create(ctx, namespace)

			// Expect(err).To(Not(HaveOccurred()))
		})

		AfterEach(func() {
			By("expect no reconcile errors occurred")
			Expect(logOutput.GetErrorCount()).To(Equal(0), logOutput.GetOutput())
			// By("Deleting the Namespace to perform the tests")
			// _ = k8sClient.Delete(ctx, namespace)
		})

		It("should successfully reconcile a custom resource for CloudflareServiceToken", func() {
			typeNamespaceName := types.NamespacedName{Name: "token1", Namespace: nsName}

			By("Creating the custom resource for the Kind CloudflareServiceToken")
			serviceToken := &v4alpha1.CloudflareServiceToken{
				ObjectMeta: metav1.ObjectMeta{
					Name:      typeNamespaceName.Name,
					Namespace: typeNamespaceName.Namespace,
				},
				Spec: v4alpha1.CloudflareServiceTokenSpec{
					Name: "servicetoken v2 test",
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
				return k8sClient.Get(ctx, expectedSecondarySecIdentity, sec)
			}).WithTimeout(20 * time.Second).WithPolling(time.Second).Should(Succeed())

			By("Make sure the status ref is what we expect")
			Eventually(func(g Gomega) { //nolint:varnamelen
				g.Expect(k8sClient.Get(ctx, typeNamespaceName, serviceToken)).ToNot(HaveOccurred())
				g.Expect(serviceToken.Status.ServiceTokenID).ToNot(BeEmpty())
				g.Expect(serviceToken.Status.SecretRef).ToNot(BeNil())
				g.Expect(serviceToken.Status.SecretRef.ClientIDKey).ToNot(BeEmpty())
				g.Expect(serviceToken.Status.SecretRef.ClientSecretKey).ToNot(BeEmpty())
				g.Expect(serviceToken.Status.SecretRef.Name).To(Equal(sec.Name))
			}).WithTimeout(10 * time.Second).WithPolling(time.Second).Should(Succeed())

			By("Checking if the resource exists in cloudflare")
			tokens, err := api.ServiceTokens(ctx)
			Expect(err).To(Not(HaveOccurred()))

			By("Renaming the secret")
			serviceToken.Spec.Name = "updated_secret_name"
			Expect(k8sClient.Update(ctx, serviceToken)).Should(Succeed())

			//
			expectedID := string(sec.Data[sec.Annotations[v4alpha1.AnnotationTokenIDKey]])
			expectedClientID := string(sec.Data[sec.Annotations[v4alpha1.AnnotationClientIDKey]])
			tokenfound := false
			for _, token := range *tokens {
				if serviceToken.Spec.Name == token.Name && expectedID == token.ID && expectedClientID == token.ClientID {
					tokenfound = true
					break
				}
			}
			Expect(tokenfound).To(BeTrue(), "token not found")
		})

		It("should successfully reconcile a custom resource for CloudflareServiceToken", func() {
			typeNamespaceName := types.NamespacedName{Name: "token2", Namespace: nsName}

			By("Creating the custom resource for the Kind CloudflareServiceToken")
			var sToken *v4alpha1.CloudflareServiceToken

			token := &v4alpha1.CloudflareServiceToken{}
			err := k8sClient.Get(ctx, typeNamespaceName, token)
			if err != nil && errors.IsNotFound(err) {
				sToken = &v4alpha1.CloudflareServiceToken{
					ObjectMeta: metav1.ObjectMeta{
						Name:      typeNamespaceName.Name,
						Namespace: typeNamespaceName.Namespace,
					},
					Spec: v4alpha1.CloudflareServiceTokenSpec{
						Name: "integration servicetoken test",
					},
				}

				err = k8sClient.Create(ctx, sToken)
				Expect(err).To(Not(HaveOccurred()))
			}

			By("Checking if the custom resource was successfully created")
			found := &v4alpha1.CloudflareServiceToken{}
			Eventually(func() error {
				return k8sClient.Get(ctx, typeNamespaceName, found)
			}).WithTimeout(10 * time.Second).WithPolling(time.Second).Should(Succeed())

			Expect(err).To(Not(HaveOccurred()))

			By("Checking to get the updated CR")
			Eventually(func() error {
				return k8sClient.Get(ctx, typeNamespaceName, found)
			}).WithTimeout(10 * time.Second).WithPolling(time.Second).Should(Succeed())

			By("Make sure the status ref is what we expect")
			Eventually(func(g Gomega) { //nolint:varnamelen
				_ = k8sClient.Get(ctx, typeNamespaceName, found)
				g.Expect(found.Status.ServiceTokenID).ToNot(BeEmpty())
				g.Expect(found.Status.SecretRef).ToNot(BeNil())
				g.Expect(found.Status.SecretRef.ClientIDKey).ToNot(BeEmpty())
				g.Expect(found.Status.SecretRef.ClientSecretKey).ToNot(BeEmpty())
				g.Expect(found.Status.SecretRef.Name).ToNot(BeEmpty())
			}).WithTimeout(10 * time.Second).WithPolling(time.Second).Should(Succeed())

			sec := &corev1.Secret{}
			By("Making sure that the secret exists")
			Eventually(func() error {
				return k8sClient.Get(ctx, typeNamespaceName, sec)
			}).WithTimeout(10 * time.Second).WithPolling(time.Second).Should(Succeed())

			By("Checking if the resource exists in cloudflare")
			tokens, err := api.ServiceTokens(ctx)
			Expect(err).To(Not(HaveOccurred()))

			secretFound := false
			for _, token := range *tokens {
				if token.Name == found.Spec.Name {
					secretFound = true

					Expect(token.ID).To(Equal(string(sec.Data[sec.Annotations[v4alpha1.AnnotationTokenIDKey]])))
					Expect(token.ClientID).To(Equal(string(sec.Data[sec.Annotations[v4alpha1.AnnotationClientIDKey]])))
				}
			}
			// we should only have 1 Token created

			Expect(secretFound).To(BeTrue(), "secret not found", found.Spec.Name, tokens)

			By("Updating the service token to move the secret")
			_ = k8sClient.Get(ctx, typeNamespaceName, found)
			found.Spec.Template.Name = "moved-secret"
			Eventually(func() error {
				return k8sClient.Update(ctx, found)
			}).WithTimeout(10 * time.Second).WithPolling(time.Second).Should(Succeed())

			By("Checking if the new secret was successfully created")
			Eventually(func() error {
				sec := &corev1.Secret{}
				return k8sClient.Get(ctx, types.NamespacedName{Name: found.Spec.Template.Name, Namespace: typeNamespaceName.Namespace}, sec)
			}).WithTimeout(10 * time.Second).WithPolling(time.Second).Should(Succeed())

			By("Make sure the status ref is what we expect")
			_ = k8sClient.Get(ctx, typeNamespaceName, found)
			Expect(found.Status.ServiceTokenID).ToNot(BeEmpty())
			Expect(found.Status.SecretRef).ToNot(BeNil())
			Expect(found.Status.SecretRef.ClientIDKey).ToNot(BeEmpty())
			Expect(found.Status.SecretRef.ClientSecretKey).ToNot(BeEmpty())
			Expect(found.Status.SecretRef.Name).ToNot(BeEmpty())

			By("Updating the secret template")
			err = k8sClient.Get(ctx, typeNamespaceName, sToken)
			Expect(err).ToNot(HaveOccurred())
			sToken.Spec.Template.Name = "moved-secret"
			sToken.Spec.Template.ClientIDKey = "keylocation"
			err = k8sClient.Update(ctx, sToken)
			Expect(err).ToNot(HaveOccurred())

			By("Make sure the status ref is what we expect")
			Eventually(func(g Gomega) {
				err = k8sClient.Get(ctx, typeNamespaceName, sToken)
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(sToken.Status.SecretRef.Name).To(Equal(sToken.Spec.Template.Name))
				g.Expect(sToken.Status.SecretRef.ClientIDKey).To(Equal(sToken.Spec.Template.ClientIDKey))
			}).Should(Succeed())

			By("Checking if the new secret was successfully created")
			Eventually(func() error {
				sec = &corev1.Secret{}
				return k8sClient.Get(ctx, types.NamespacedName{Name: sToken.Spec.Template.Name, Namespace: typeNamespaceName.Namespace}, sec)
			}).WithTimeout(10 * time.Second).WithPolling(time.Second).Should(Succeed())
			Expect(sec.Data).To(HaveKey(sToken.Spec.Template.ClientIDKey))

			By("Checking if the old secret was removed")
			Eventually(func() error {
				sec := &corev1.Secret{}
				return k8sClient.Get(ctx, types.NamespacedName{Name: typeNamespaceName.Name, Namespace: typeNamespaceName.Namespace}, sec)
			}).WithTimeout(10 * time.Second).WithPolling(time.Second).ShouldNot(Succeed())
		})

		It("should successfully allow removal of resource if externally deleted", func() {
			typeNamespaceName := types.NamespacedName{Name: "token4", Namespace: nsName}

			By("Creating the custom resource for the Kind CloudflareServiceToken")
			token := &v4alpha1.CloudflareServiceToken{}

			err := k8sClient.Get(ctx, typeNamespaceName, token)
			if err != nil && errors.IsNotFound(err) {
				token = &v4alpha1.CloudflareServiceToken{
					ObjectMeta: metav1.ObjectMeta{
						Name:      typeNamespaceName.Name,
						Namespace: typeNamespaceName.Namespace,
					},
					Spec: v4alpha1.CloudflareServiceTokenSpec{
						Name: "integration servicetoken test4",
					},
				}

				err = k8sClient.Create(ctx, token)
				Expect(err).To(Not(HaveOccurred()))
			}

			By("Checking to get the updated CR")
			Eventually(func() error {
				return k8sClient.Get(ctx, typeNamespaceName, token)
			}).WithTimeout(10 * time.Second).WithPolling(time.Second).Should(Succeed())

			By("Make sure the status ref is what we expect")
			Eventually(func(g Gomega) {
				_ = k8sClient.Get(ctx, typeNamespaceName, token)
				g.Expect(token.Status.ServiceTokenID).ToNot(BeEmpty())
			}).WithTimeout(10 * time.Second).WithPolling(time.Second).Should(Succeed())

			By("Externally removing the token")
			Expect(api.DeleteAccessServiceToken(ctx, token.Status.ServiceTokenID)).To(Succeed())

			By("Removing the access service token")
			Expect(k8sClient.Delete(ctx, token)).To(Succeed())
		})

		It("should successfully reconcile a custom resource for CloudflareServiceToken", func() {
			typeNamespaceName := types.NamespacedName{Name: "token3", Namespace: nsName}

			By("Creating the custom resource for the Kind CloudflareServiceToken")
			token := &v4alpha1.CloudflareServiceToken{}

			err := k8sClient.Get(ctx, typeNamespaceName, token)
			if err != nil && errors.IsNotFound(err) {
				token = &v4alpha1.CloudflareServiceToken{
					ObjectMeta: metav1.ObjectMeta{
						Name:      typeNamespaceName.Name,
						Namespace: typeNamespaceName.Namespace,
					},
					Spec: v4alpha1.CloudflareServiceTokenSpec{
						Name: "integration servicetoken test3",
					},
				}

				err = k8sClient.Create(ctx, token)
				Expect(err).To(Not(HaveOccurred()))
			}

			By("Checking to get the updated CR")
			Eventually(func() error {
				return k8sClient.Get(ctx, typeNamespaceName, token)
			}).WithTimeout(10 * time.Second).WithPolling(time.Second).Should(Succeed())

			By("Make sure the annotation is not present")
			Eventually(func(g Gomega) {
				_ = k8sClient.Get(ctx, typeNamespaceName, token)
				g.Expect(token.Status.ServiceTokenID).ToNot(BeEmpty())
				keyValue, keyExists := token.Annotations[v4alpha1.AnnotationPreventDestroy]
				g.Expect(keyExists).To(BeFalse())
				g.Expect(keyValue).To(Or(BeEmpty(), Equal("false")))
			}).WithTimeout(10 * time.Second).WithPolling(time.Second).Should(Succeed())

			By("Make sure the service token exists on cloudflare")
			tokens, err := api.ServiceTokens(ctx)
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
				tokens, _ := api.ServiceTokens(ctx)
				var foundToken *cftypes.ExtendedServiceToken
				for _, cfToken := range *tokens {
					if cfToken.ID == token.Status.ServiceTokenID {
						foundToken = &cfToken
					}
				}

				g.Expect(foundToken).To(BeNil())
			}).WithTimeout(10 * time.Second).WithPolling(time.Second).Should(Succeed())
		})

		It("should successfully not remove the resource in CF if annotation is set", func() {
			typeNamespaceName := types.NamespacedName{Name: "token5", Namespace: nsName}

			By("Creating the custom resource for the Kind CloudflareServiceToken")
			token := &v4alpha1.CloudflareServiceToken{}

			err := k8sClient.Get(ctx, typeNamespaceName, token)
			if err != nil && errors.IsNotFound(err) {
				token = &v4alpha1.CloudflareServiceToken{
					ObjectMeta: metav1.ObjectMeta{
						Name:      typeNamespaceName.Name,
						Namespace: typeNamespaceName.Namespace,
						Annotations: map[string]string{
							v4alpha1.AnnotationPreventDestroy: "true",
						},
					},
					Spec: v4alpha1.CloudflareServiceTokenSpec{
						Name: "integration servicetoken test5",
					},
				}

				err = k8sClient.Create(ctx, token)
				Expect(err).To(Not(HaveOccurred()))
			}

			By("Checking to get the updated CR")
			Eventually(func() error {
				return k8sClient.Get(ctx, typeNamespaceName, token)
			}).WithTimeout(10 * time.Second).WithPolling(time.Second).Should(Succeed())

			By("Make sure the status ref is what we expect")
			Eventually(func(g Gomega) {
				_ = k8sClient.Get(ctx, typeNamespaceName, token)
				g.Expect(token.Status.ServiceTokenID).ToNot(BeEmpty())
			}).WithTimeout(10 * time.Second).WithPolling(time.Second).Should(Succeed())

			By("Make sure the service token exists on cloudflare")
			tokens, err := api.ServiceTokens(ctx)
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
			tokens, err = api.ServiceTokens(ctx)
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
