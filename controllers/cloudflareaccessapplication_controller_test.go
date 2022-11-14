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

var _ = Describe("CloudflareAccessApplication controller", Ordered, func() {
	BeforeAll(func() {
		ctx := context.Background()

		By("Removing all existing access apps")
		apps, err := api.AccessApplications(ctx)
		Expect(err).To(Not(HaveOccurred()))
		for _, app := range apps {
			err = api.DeleteAccessApplication(ctx, app.ID)
			Expect(err).To(Not(HaveOccurred()))
		}
	})

	Context("CloudflareAccessApplication controller test", func() {

		const cloudflareName = "cloudflare-app"

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
			err := k8sClient.Create(ctx, namespace)
			Expect(err).To(Not(HaveOccurred()))
		})

		AfterEach(func() {
			By("Deleting the Namespace to perform the tests")
			_ = k8sClient.Delete(ctx, namespace)
		})

		It("should successfully reconcile a custom resource for CloudflareAccessApplication", func() {
			By("Creating the custom resource for the Kind CloudflareAccessApplication")
			group := &v1alpha1.CloudflareAccessApplication{}
			err := k8sClient.Get(ctx, typeNamespaceName, group)
			if err != nil && errors.IsNotFound(err) {
				// Let's mock our custom resource at the same way that we would
				// apply on the cluster the manifest under config/samples
				group := &v1alpha1.CloudflareAccessApplication{
					ObjectMeta: metav1.ObjectMeta{
						Name:      cloudflareName,
						Namespace: namespace.Name,
					},
					Spec: v1alpha1.CloudflareAccessApplicationSpec{
						Name:   "integration test",
						Domain: "integration.cf-operator-tests.uk",
					},
				}

				err = k8sClient.Create(ctx, group)
				Expect(err).To(Not(HaveOccurred()))
			}

			By("Checking if the custom resource was successfully created")
			Eventually(func() error {
				found := &v1alpha1.CloudflareAccessApplication{}
				return k8sClient.Get(ctx, typeNamespaceName, found)
			}, time.Minute, time.Second).Should(Succeed())

			By("Reconciling the custom resource created")
			accessGroupReconciler := &CloudflareAccessApplicationReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			_, err = accessGroupReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespaceName,
			})
			Expect(err).To(Not(HaveOccurred()))

			found := &v1alpha1.CloudflareAccessApplication{}
			By("Checking the latest Status should have the ID of the resource")
			Eventually(func() string {
				found = &v1alpha1.CloudflareAccessApplication{}
				k8sClient.Get(ctx, typeNamespaceName, found)
				return found.Status.AccessApplicationID
			}, time.Minute, time.Second).Should(Not(BeEmpty()))

			By("Cloudflare resource should equal the spec")
			cfResource, err := api.AccessApplication(ctx, found.Status.AccessApplicationID)
			Expect(err).To(Not(HaveOccurred()))
			Expect(cfResource.Name).To(Equal(cfResource.Name))
		})
	})
})
