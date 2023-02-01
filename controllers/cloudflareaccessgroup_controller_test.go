//go:build integration

package controllers

import (
	"context"
	"time"

	v1alpha1 "github.com/bojanzelic/cloudflare-zero-trust-operator/api/v1alpha1"
	"github.com/bojanzelic/cloudflare-zero-trust-operator/internal/cfapi"
	"github.com/cloudflare/cloudflare-go"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("CloudflareAccessGroup controller", Ordered, func() {
	BeforeAll(func() {
		ctx := context.Background()

		By("Removing all existing access groups")
		groups, err := api.AccessGroups(ctx)
		Expect(err).To(Not(HaveOccurred()))
		for _, group := range groups {
			_ = api.DeleteAccessGroup(ctx, group.ID)
			//Expect(err).To(Not(HaveOccurred()))
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
			logOutput.Clear()

			By("Creating the Namespace to perform the tests")
			k8sClient.Create(ctx, namespace)

			//Expect(err).To(Not(HaveOccurred()))
		})

		AfterEach(func() {
			By("expect no reconcile errors occured")
			Expect(logOutput.GetErrorCount()).To(Equal(0), logOutput.GetOutput())
			// By("Deleting the Namespace to perform the tests")
			// _ = k8sClient.Delete(ctx, namespace)
		})

		It("should successfully reconcile if a CloudflareAccessGroup AlreadyExists", func() {
			By("Pre-creating a cloudflare access group")

			ag, err := api.CreateAccessGroup(ctx, cloudflare.AccessGroup{
				Name: "existing-access-group",
				Include: []interface{}{
					cfapi.NewAccessGroupEmail("test1@cf-operator-tests.uk"),
				},
			})
			Expect(err).ToNot(HaveOccurred())

			By("Creating the same custom resource for the Kind CloudflareAccessGroup")
			group := &v1alpha1.CloudflareAccessGroup{
				ObjectMeta: metav1.ObjectMeta{
					Name:      ag.Name,
					Namespace: namespace.Name,
				},
				Spec: v1alpha1.CloudflareAccessGroupSpec{
					Name: ag.Name,
					Include: []v1alpha1.CloudFlareAccessGroupRule{
						{
							Emails: []string{"test2@cf-operator-tests.uk"},
						},
					},
				},
			}

			err = k8sClient.Create(ctx, group)
			Expect(err).To(Not(HaveOccurred()))

			found := &v1alpha1.CloudflareAccessGroup{}
			By("Checking the latest Status should have the ID of the resource")
			Eventually(func() string {
				found = &v1alpha1.CloudflareAccessGroup{}
				k8sClient.Get(ctx, types.NamespacedName{Name: group.Name, Namespace: group.Namespace}, found)
				return found.Status.AccessGroupID
			}, time.Second*10, time.Second).Should(Equal(ag.ID))
		})

		It("should successfully reconcile a custom resource for CloudflareAccessGroup", func() {
			By("Creating the custom resource for the Kind CloudflareAccessGroup")
			group := &v1alpha1.CloudflareAccessGroup{
				ObjectMeta: metav1.ObjectMeta{
					Name:      cloudflareName,
					Namespace: namespace.Name,
				},
				Spec: v1alpha1.CloudflareAccessGroupSpec{
					Name: "integration accessgroup test",
					Include: []v1alpha1.CloudFlareAccessGroupRule{
						{
							Emails: []string{"test@cf-operator-tests.uk"},
						},
					},
				},
			}

			err := k8sClient.Create(ctx, group)
			Expect(err).To(Not(HaveOccurred()))

			By("Checking if the custom resource was successfully created")
			Eventually(func() error {
				found := &v1alpha1.CloudflareAccessGroup{}
				return k8sClient.Get(ctx, typeNamespaceName, found)
			}, time.Minute, time.Second).Should(Succeed())

			found := &v1alpha1.CloudflareAccessGroup{}
			By("Checking the latest Status should have the ID of the resource")
			Eventually(func() string {
				found = &v1alpha1.CloudflareAccessGroup{}
				k8sClient.Get(ctx, typeNamespaceName, found)
				return found.Status.AccessGroupID
			}, time.Minute, time.Second).Should(Not(BeEmpty()))

			By("Cloudflare resource should equal the spec")
			cfResource, err := api.AccessGroup(ctx, found.Status.AccessGroupID)
			Expect(err).To(Not(HaveOccurred()))
			Expect(cfResource.Name).To(Equal(found.Spec.Name))

			By("Updating the name of the resource")
			found.Spec.Name = "updated name"
			k8sClient.Update(ctx, found)
			Expect(err).To(Not(HaveOccurred()))

			By("Cloudflare resource should equal the updated spec")
			Eventually(func() string {
				cfResource, err = api.AccessGroup(ctx, found.Status.AccessGroupID)
				return cfResource.Name

			}, time.Minute, time.Second).Should(Equal(found.Spec.Name))
		})
	})
})
