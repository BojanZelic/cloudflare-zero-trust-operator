//go:build integration
// +build integration

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

var _ = Describe("CloudflareAccessGroup controller", Ordered, func() {
	BeforeAll(func() {
		ctx := context.Background()

		By("Removing all existing access groups")
		groups, err := api.AccessGroups(ctx)
		Expect(err).To(Not(HaveOccurred()))
		for _, group := range groups {
			err = api.DeleteAccessGroup(ctx, group.ID)
			Expect(err).To(Not(HaveOccurred()))
		}
	})

	Context("CloudflareAccessGroup controller test", func() {

		const cloudflareName = "test-cloudflare"

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

		It("should successfully reconcile a custom resource for Memcached", func() {
			By("Creating the custom resource for the Kind Memcached")
			group := &v1alpha1.CloudflareAccessGroup{}
			err := k8sClient.Get(ctx, typeNamespaceName, group)
			if err != nil && errors.IsNotFound(err) {
				// Let's mock our custom resource at the same way that we would
				// apply on the cluster the manifest under config/samples
				group := &v1alpha1.CloudflareAccessGroup{
					ObjectMeta: metav1.ObjectMeta{
						Name:      cloudflareName,
						Namespace: namespace.Name,
					},
					Spec: v1alpha1.CloudflareAccessGroupSpec{
						Name: "integration test",
						Include: []v1alpha1.CloudFlareAccessGroupRule{
							{
								Emails: []string{"test@cf-operator-tests.uk"},
							},
						},
					},
				}

				err = k8sClient.Create(ctx, group)
				Expect(err).To(Not(HaveOccurred()))
			}

			By("Checking if the custom resource was successfully created")
			Eventually(func() error {
				found := &v1alpha1.CloudflareAccessGroup{}
				return k8sClient.Get(ctx, typeNamespaceName, found)
			}, time.Minute, time.Second).Should(Succeed())

			By("Reconciling the custom resource created")
			accessGroupReconciler := &CloudflareAccessGroupReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			_, err = accessGroupReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespaceName,
			})
			Expect(err).To(Not(HaveOccurred()))

			// By("Checking if Deployment was successfully created in the reconciliation")
			// Eventually(func() error {
			// 	found := &appsv1.Deployment{}
			// 	return k8sClient.Get(ctx, typeNamespaceName, found)
			// }, time.Minute, time.Second).Should(Succeed())

			// By("Checking the latest Status Condition added to the Memcached instance")
			// Eventually(func() error {
			// 	if memcached.Status.Conditions != nil && len(memcached.Status.Conditions) != 0 {
			// 		latestStatusCondition := memcached.Status.Conditions[len(memcached.Status.Conditions)-1]
			// 		expectedLatestStatusCondition := metav1.Condition{Type: typeAvailableMemcached,
			// 			Status: metav1.ConditionTrue, Reason: "Reconciling",
			// 			Message: fmt.Sprintf("Deployment for custom resource (%s) with %d replicas created successfully", memcached.Name, memcached.Spec.Size)}
			// 		if latestStatusCondition != expectedLatestStatusCondition {
			// 			return fmt.Errorf("The latest status condition added to the memcached instance is not as expected")
			// 		}
			// 	}
			// 	return nil
			// }, time.Minute, time.Second).Should(Succeed())
		})
	})
})
