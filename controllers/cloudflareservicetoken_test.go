//go:build integration

package controllers

import (
	"context"
	"time"

	v1alpha1 "github.com/bojanzelic/cloudflare-zero-trust-operator/api/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var _ = FDescribe("CloudflareServiceToken controller", Ordered, func() {
	BeforeAll(func() {
		ctx := context.Background()

		By("Removing all existing service tokens")
		tokens, err := api.ServiceTokens(ctx)
		Expect(err).To(Not(HaveOccurred()))
		for _, token := range tokens {
			err = api.DeleteAccessServiceToken(ctx, token.ID)
			Expect(err).To(Not(HaveOccurred()))
		}
	})

	Context("CloudflareServiceToken controller test", func() {

		const cloudflareName = "stokens-cloudflare"

		ctx := context.Background()

		namespace := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name:      cloudflareName,
				Namespace: cloudflareName,
			},
		}

		typeNamespaceName := types.NamespacedName{Name: cloudflareName, Namespace: cloudflareName}

		BeforeEach(func() {
			By("Creating the Namespace to perform the tests")
			k8sClient.Create(ctx, namespace)
			//Expect(err).To(Not(HaveOccurred()))
		})

		// AfterEach(func() {
		// 	By("Deleting the Namespace to perform the tests")
		// 	_ = k8sClient.Delete(ctx, namespace)
		// })

		It("should successfully reconcile a custom resource for CloudflareServiceToken", func() {
			By("Creating the custom resource for the Kind CloudflareServiceToken")
			token := &v1alpha1.CloudflareServiceToken{}
			err := k8sClient.Get(ctx, typeNamespaceName, token)
			if err != nil && errors.IsNotFound(err) {
				group := &v1alpha1.CloudflareServiceToken{
					ObjectMeta: metav1.ObjectMeta{
						Name:      typeNamespaceName.Name,
						Namespace: typeNamespaceName.Namespace,
					},
					Spec: v1alpha1.CloudflareServiceTokenSpec{
						Name: "integration test",
					},
				}

				err = k8sClient.Create(ctx, group)
				Expect(err).To(Not(HaveOccurred()))
			}

			By("Checking if the custom resource was successfully created")
			found := &v1alpha1.CloudflareServiceToken{}
			Eventually(func() error {
				return k8sClient.Get(ctx, typeNamespaceName, found)
			}, time.Second*10, time.Second).Should(Succeed())

			By("Reconciling the custom resource created")
			serviceTokenReconciler := &CloudflareServiceTokenReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			_, err = serviceTokenReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespaceName,
			})
			Expect(err).To(Not(HaveOccurred()))

			By("Checking if the secret was successfully created")
			Eventually(func() error {
				found := &corev1.Secret{}
				return k8sClient.Get(ctx, typeNamespaceName, found)
			}, time.Second*10, time.Second).Should(Succeed())

			By("Updating the service token to move the secret")
			k8sClient.Get(ctx, typeNamespaceName, found)
			found.Spec.Template.Name = "moved-secret"
			Eventually(func() error {
				return k8sClient.Update(ctx, found)
			}, time.Second*10, time.Second).Should(Succeed())

			By("Reconciling the updated resource created")
			_, err = serviceTokenReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespaceName,
			})
			Expect(err).To(Not(HaveOccurred()))

			By("Checking if the new secret was successfully created")
			Eventually(func() error {
				sec := &corev1.Secret{}
				return k8sClient.Get(ctx, types.NamespacedName{Name: found.Spec.Template.Name, Namespace: typeNamespaceName.Namespace}, sec)
			}, time.Second*10, time.Second).Should(Succeed())

			By("Make sure the status ref is what we expect")
			k8sClient.Get(ctx, typeNamespaceName, found)
			Expect(found.Status.ServiceTokenID).ToNot(BeEmpty())
			Expect(found.Status.SecretRef).ToNot(BeNil())
			Expect(found.Status.SecretRef.ClientIDKey).ToNot(BeEmpty())
			Expect(found.Status.SecretRef.ClientSecretKey).ToNot(BeEmpty())
			Expect(found.Status.SecretRef.Name).ToNot(BeEmpty())

			// By("Reconciling the custom resource created")
			// accessGroupReconciler := &CloudflareServiceTokenReconciler{
			// 	Client: k8sClient,
			// 	Scheme: k8sClient.Scheme(),
			// }

			// _, err = accessGroupReconciler.Reconcile(ctx, reconcile.Request{
			// 	NamespacedName: typeNamespaceName,
			// })
			// Expect(err).To(Not(HaveOccurred()))

			// found := &v1alpha1.CloudflareServiceToken{}
			// By("Checking the latest Status should have the ID of the resource")
			// Eventually(func() string {
			// 	found = &v1alpha1.CloudflareServiceToken{}
			// 	k8sClient.Get(ctx, typeNamespaceName, found)
			// 	return found.Status.AccessGroupID
			// }, time.Minute, time.Second).Should(Not(BeEmpty()))

			// By("Cloudflare resource should equal the spec")
			// cfResource, err := api.AccessGroup(ctx, found.Status.AccessGroupID)
			// Expect(err).To(Not(HaveOccurred()))
			// Expect(cfResource.Name).To(Equal(found.Spec.Name))

			// By("Updating the name of the resource")
			// found.Spec.Name = "updated name"
			// k8sClient.Update(ctx, found)
			// Expect(err).To(Not(HaveOccurred()))

			// By("Reconciling the updated resource")
			// _, err = accessGroupReconciler.Reconcile(ctx, reconcile.Request{
			// 	NamespacedName: typeNamespaceName,
			// })
			// Expect(err).To(Not(HaveOccurred()))

			// By("Cloudflare resource should equal the updated spec")
			// cfResource, err = api.AccessGroup(ctx, found.Status.AccessGroupID)
			// Expect(err).To(Not(HaveOccurred()))
			// Expect(cfResource.Name).To(Equal(found.Spec.Name))
		})
	})
})
