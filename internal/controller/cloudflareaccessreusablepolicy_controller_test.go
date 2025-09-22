//go:build integration

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

var _ = Describe("CloudflareAccessReusablePolicy controller", Ordered, func() {
	BeforeAll(func() { insertedTracer.ResetStores() })
	AfterAll(func() {
		errs := insertedTracer.UninstallFromCF(api)
		Expect(errs).To(BeEmpty())
	})

	//
	//
	//

	Context("CloudflareAccessReusablePolicy controller test", func() {
		const testScopedNamespace = "zto-testing-arp"

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

		It("should fail to reconcile CloudflareAccessReusablePolicy policies with bad references", func() {
			By("Creating the custom resource for the Kind CloudflareAccessReusablePolicy")
			arpNN := types.NamespacedName{Name: "test-1-arp", Namespace: testScopedNamespace}
			arp := &v4alpha1.CloudflareAccessReusablePolicy{
				ObjectMeta: metav1.ObjectMeta{
					Name:      arpNN.Name,
					Namespace: arpNN.Namespace,
				},
				Spec: v4alpha1.CloudflareAccessReusablePolicySpec{
					Name: "ZTO AccessReusablePolicy Tests - 1 - Policy",
					Include: v4alpha1.CloudFlareAccessRules{
						AccessGroupRefs: []string{
							"inanynamespace/idontexist",
						},
					},
				},
			}
			Expect(k8sClient.Create(ctx, arp)).ToNot(HaveOccurred())

			//
			ByExpectingCFResourceTo_NOT_BeReady(ctx, arp).Should(Succeed())

			//
			By("Removing dangling resource to prevent looping on failure")
			Expect(k8sClient.Delete(ctx, arp)).ToNot(HaveOccurred())

			//
			ByExpectingDeletionOf(arp).Should(Succeed())
		})

		It("should successfully reconcile CloudflareAccessReusablePolicy policies with references", func() {
			By("pre-create an access group")
			groupNN := types.NamespacedName{Name: "test-2-group", Namespace: testScopedNamespace}
			group := &v4alpha1.CloudflareAccessGroup{
				ObjectMeta: metav1.ObjectMeta{
					Name:      groupNN.Name,
					Namespace: groupNN.Namespace,
				},
				Spec: v4alpha1.CloudflareAccessGroupSpec{
					Name: "ZTO AccessReusablePolicy Tests - 2 - Group",
					Include: v4alpha1.CloudFlareAccessRules{
						Emails: []string{
							produceOwnedEmail("zto-test-arp-2"),
						},
					},
				},
			}
			Expect(k8sClient.Create(ctx, group)).ToNot(HaveOccurred())

			//
			ByExpectingCFResourceToBeReady(ctx, group).Should(Succeed())

			//
			//
			//

			By("pre-create a service token")
			tokenNN := types.NamespacedName{Name: "test-2-stoken", Namespace: testScopedNamespace}
			token := &v4alpha1.CloudflareServiceToken{
				ObjectMeta: metav1.ObjectMeta{
					Name:      tokenNN.Name,
					Namespace: tokenNN.Namespace,
				},
				Spec: v4alpha1.CloudflareServiceTokenSpec{
					Name: "ZTO AccessReusablePolicy Tests - 2 - SToken",
				},
			}
			Expect(k8sClient.Create(ctx, token)).ToNot(HaveOccurred())

			//
			ByExpectingCFResourceToBeReady(ctx, token).Should(Succeed())

			//
			//
			//

			By("Creating the custom resource for the Kind CloudflareAccessReusablePolicy")
			arpNN := types.NamespacedName{Name: "test-2-arp", Namespace: testScopedNamespace}
			arp := &v4alpha1.CloudflareAccessReusablePolicy{
				ObjectMeta: metav1.ObjectMeta{
					Name:      arpNN.Name,
					Namespace: arpNN.Namespace,
				},
				Spec: v4alpha1.CloudflareAccessReusablePolicySpec{
					Name: "ZTO AccessReusablePolicy Tests - 2 - Policy",
					Include: v4alpha1.CloudFlareAccessRules{
						AccessGroupRefs: []string{
							v4alpha1.ParsedNamespacedName(types.NamespacedName{Name: group.Name, Namespace: group.Namespace}),
						},
						ServiceTokenRefs: []string{
							v4alpha1.ParsedNamespacedName(types.NamespacedName{Name: token.Name, Namespace: token.Namespace}),
						},
					},
				},
			}
			Expect(k8sClient.Create(ctx, arp)).ToNot(HaveOccurred())

			//
			ByExpectingCFResourceToBeReady(ctx, arp).Should(Succeed())

			//
			//
			//

			By("Creating the custom resource for the Kind CloudflareAccessApplication")
			appNN := types.NamespacedName{Name: "test-2-app", Namespace: testScopedNamespace}
			app := &v4alpha1.CloudflareAccessApplication{
				ObjectMeta: metav1.ObjectMeta{
					Name:      appNN.Name,
					Namespace: appNN.Namespace,
				},
				Spec: v4alpha1.CloudflareAccessApplicationSpec{
					Name:   "ZTO AccessReusablePolicy Tests - 2 - App",
					Domain: produceOwnedFQDN("zto-test-arp-2"),
					PolicyRefs: []string{
						v4alpha1.ParsedNamespacedName(arpNN),
					},
				},
			}
			Expect(k8sClient.Create(ctx, app)).ToNot(HaveOccurred())

			//
			ByExpectingCFResourceToBeReady(ctx, app).Should(Succeed())
		})

		It("should successfully reconcile a custom resource for CloudflareAccessReusablePolicy", func() {
			By("Creating the custom resource for the Kind CloudflareAccessReusablePolicy")
			arpNN := types.NamespacedName{Name: "test-3-arp", Namespace: testScopedNamespace}
			arp := &v4alpha1.CloudflareAccessReusablePolicy{
				ObjectMeta: metav1.ObjectMeta{
					Name:      arpNN.Name,
					Namespace: arpNN.Namespace,
				},
				Spec: v4alpha1.CloudflareAccessReusablePolicySpec{
					Name: "ZTO AccessReusablePolicy Tests - 3 - Policy",
					Include: v4alpha1.CloudFlareAccessRules{
						EmailDomains: []string{
							produceOwnedFQDN("zto-test-arp-3"),
						},
					},
				},
			}
			Expect(k8sClient.Create(ctx, arp)).ToNot(HaveOccurred())

			//
			ByExpectingCFResourceToBeReady(ctx, arp).Should(Succeed())

			By("Cloudflare resource should equal the spec")
			cfResource, err := api.AccessReusablePolicy(ctx, arp.GetCloudflareUUID())
			Expect(err).ToNot(HaveOccurred())
			Expect(cfResource.Name).To(Equal(arp.Spec.Name))

			By("Updating the name of the resource")

			addDirtyingSuffix(&arp.Spec.Name)
			Expect(k8sClient.Update(ctx, arp)).ToNot(HaveOccurred())

			// Await for resource to be ready again
			ByExpectingCFResourceToBeReady(ctx, arp).Should(Succeed())

			By("Cloudflare resource should equal the updated spec")
			cfResource, err = api.AccessReusablePolicy(ctx, arp.GetCloudflareUUID())
			Expect(err).ToNot(HaveOccurred())
			Expect(cfResource.Name).To(Equal(arp.Spec.Name))

			By("Cloudflare resource should be deleted")
			Expect(k8sClient.Delete(ctx, arp)).ToNot(HaveOccurred())

			//
			ByExpectingDeletionOf(arp).Should(Succeed())
		})

		It("should successfully reconcile CloudflareAccessReusablePolicy whose AccessReusablePolicyID references a missing Reusable Policy", func() {
			By("Recreating the custom resource for the Kind CloudflareAccessReusablePolicy")
			arpNN := types.NamespacedName{Name: "test-4-arp", Namespace: testScopedNamespace}
			arp := &v4alpha1.CloudflareAccessReusablePolicy{
				ObjectMeta: metav1.ObjectMeta{
					Name:      arpNN.Name,
					Namespace: arpNN.Namespace,
				},
				Spec: v4alpha1.CloudflareAccessReusablePolicySpec{
					Name: "ZTO AccessReusablePolicy Tests - 4 - Policy",
					Include: v4alpha1.CloudFlareAccessRules{
						EmailDomains: []string{
							produceOwnedFQDN("zto-test-arp-4"),
						},
					},
				},
			}
			Expect(k8sClient.Create(ctx, arp)).ToNot(HaveOccurred())

			//
			ByExpectingCFResourceToBeReady(ctx, arp).Should(Succeed())

			By("Delete associated CF Application")
			oldAccessReusablePolicyID := arp.GetCloudflareUUID()
			Expect(api.DeleteAccessReusablePolicy(ctx, arp.GetCloudflareUUID())).ToNot(HaveOccurred())

			By("re-trigger reconcile by updating access application")
			addDirtyingSuffix(&arp.Spec.Name)
			Expect(k8sClient.Update(ctx, arp)).ToNot(HaveOccurred())

			// Await for resource to be ready again
			ByExpectingCFResourceToBeReady(ctx, arp).Should(Succeed())
			Expect(arp.GetCloudflareUUID()).ToNot(Equal(oldAccessReusablePolicyID))

			By("Cloudflare resource should equal the updated spec")
			cfResource, err := api.AccessReusablePolicy(ctx, arp.GetCloudflareUUID())
			Expect(err).ToNot(HaveOccurred())
			Expect(cfResource.Name).To(Equal(arp.Spec.Name))
		})
	})
})
