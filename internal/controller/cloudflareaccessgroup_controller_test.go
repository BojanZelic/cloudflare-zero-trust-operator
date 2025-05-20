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

var _ = Describe("CloudflareAccessGroup controller", Ordered, func() {
	BeforeAll(func() { insertedTracer.ResetCFUUIDs() })
	AfterAll(func() { insertedTracer.UninstallFromCF(api) })

	//
	//
	//

	const testScopedNamespace = "zto-testing-group"

	BeforeEach(func() {
		ctrlErrors.Clear()
	})
	AfterEach(func() {
		// By("expect no reconcile errors occurred")
		// Expect(ctrlErrors).To(BeEmpty())
	})

	Context("CloudflareAccessGroup controller test - single namespace", func() {
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

		//
		//
		//

		It("should successfully reconcile if a CloudflareAccessGroup already exists with the same name", func() {
			By("Pre-creating a cloudflare access group")
			ag, err := api.CreateAccessGroup(ctx, &v4alpha1.CloudflareAccessGroup{ //nolint:varnamelen
				Spec: v4alpha1.CloudflareAccessGroupSpec{
					Name: "ZTO AccessGroup Tests - 1 - Group",
					Include: v4alpha1.CloudFlareAccessRules{
						Emails: []string{"test1@cf-operator-tests.uk"},
					},
				},
			})
			Expect(err).ToNot(HaveOccurred())

			By("Creating the same custom resource for the Kind CloudflareAccessGroup")
			groupNN := types.NamespacedName{Name: "test-1-group", Namespace: testScopedNamespace}
			group := &v4alpha1.CloudflareAccessGroup{
				ObjectMeta: metav1.ObjectMeta{
					Name:      groupNN.Name,
					Namespace: groupNN.Namespace,
				},
				Spec: v4alpha1.CloudflareAccessGroupSpec{
					Name: ag.Name, // same name !
					Include: v4alpha1.CloudFlareAccessRules{
						Emails: []string{"test2@cf-operator-tests.uk"},
					},
				},
			}
			Expect(k8sClient.Create(ctx, group)).To(Not(HaveOccurred()))

			//
			ByExpectingCFResourceToBeReady(ctx,
				groupNN,
				&v4alpha1.CloudflareAccessGroup{},
			).Should(Succeed())
		})

		It("should successfully reconcile a custom resource for CloudflareAccessGroup", func() {
			By("Creating the custom resource for the Kind CloudflareAccessGroup")
			groupNN := types.NamespacedName{Name: "test-2-group", Namespace: testScopedNamespace}
			group := &v4alpha1.CloudflareAccessGroup{
				ObjectMeta: metav1.ObjectMeta{
					Name:      groupNN.Name,
					Namespace: groupNN.Namespace,
				},
				Spec: v4alpha1.CloudflareAccessGroupSpec{
					Name: "ZTO AccessGroup Tests - 2 - Group",
					Include: v4alpha1.CloudFlareAccessRules{
						Emails: []string{"test@cf-operator-tests.uk"},
					},
				},
			}
			Expect(k8sClient.Create(ctx, group)).To(Not(HaveOccurred()))

			//
			foundGroup := v4alpha1.CloudflareAccessGroup{}
			ByExpectingCFResourceToBeReady(ctx,
				groupNN,
				&foundGroup,
			).Should(Succeed())

			By("Cloudflare resource should equal the spec")
			cfResource, err := api.AccessGroup(ctx, foundGroup.GetCloudflareUUID())
			Expect(err).To(Not(HaveOccurred()))
			Expect(cfResource.Name).To(Equal(foundGroup.Spec.Name))

			By("Updating the name of the resource")
			setUpdtdName(&foundGroup.Spec.Name)
			Expect(k8sClient.Update(ctx, &foundGroup)).To(Not(HaveOccurred()))

			//
			ByExpectingCFResourceToBeReady(ctx,
				groupNN,
				&foundGroup,
			).Should(Succeed())

			By("Cloudflare resource should equal the updated spec")
			cfResource, err = api.AccessGroup(ctx, foundGroup.GetCloudflareUUID())
			Expect(err).To(Not(HaveOccurred()))
			Expect(cfResource.Name).To(Equal(foundGroup.Spec.Name))
		})

		It("should successfully reconcile CloudflareAccessApplication policies with references", func() {
			By("pre-create a service token")
			sTokenNN := types.NamespacedName{Name: "test-3-stoken", Namespace: testScopedNamespace}
			token := &v4alpha1.CloudflareServiceToken{
				ObjectMeta: metav1.ObjectMeta{
					Name:      sTokenNN.Name,
					Namespace: sTokenNN.Namespace,
				},
				Spec: v4alpha1.CloudflareServiceTokenSpec{
					Name: "ZTO AccessGroup Tests - 3 - Service Token",
				},
			}
			Expect(k8sClient.Create(ctx, token)).To(Not(HaveOccurred()))

			//
			ByExpectingCFResourceToBeReady(ctx,
				sTokenNN,
				token,
			).Should(Succeed())

			//
			By("Creating access group")
			groupNN := types.NamespacedName{Name: "test-3-group", Namespace: testScopedNamespace}
			group := &v4alpha1.CloudflareAccessGroup{
				ObjectMeta: metav1.ObjectMeta{
					Name:      groupNN.Name,
					Namespace: groupNN.Namespace,
				},
				Spec: v4alpha1.CloudflareAccessGroupSpec{
					Name: "ZTO AccessGroup Tests - 3 - Group",
					Include: v4alpha1.CloudFlareAccessRules{
						ServiceTokenRefs: []string{
							v4alpha1.ParsedNamespacedName(sTokenNN),
						},
					},
				},
			}
			Expect(k8sClient.Create(ctx, group)).To(Not(HaveOccurred()))

			//
			ByExpectingCFResourceToBeReady(ctx,
				groupNN,
				group,
			).Should(Succeed())
		})
	})

	Context("CloudflareAccessGroup controller test - multiple namespaces", func() {
		//
		ctx := context.Background()
		oneNS := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: testScopedNamespace + "-1",
			},
		}
		twoNS := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: testScopedNamespace + "-2",
			},
		}

		BeforeAll(func() {
			By("Creating the Namespaces to perform the tests")
			_ = k8sClient.Create(ctx, oneNS)
			_ = k8sClient.Create(ctx, twoNS)
		})
		AfterAll(func() {
			By("Deleting the Namespaces to perform the tests")
			_ = k8sClient.Delete(ctx, oneNS)
			_ = k8sClient.Delete(ctx, twoNS)
			// ignore error because of https://book.kubebuilder.io/reference/envtest.html#namespace-usage-limitation
			// Expect(err).To(Not(HaveOccurred()))
		})

		It("should successfully reconcile CloudflareAccessApplication policies with references, from another namespace", func() {
			By("pre-create a service token in namespace one")
			sTokenNN := types.NamespacedName{Name: "test-4-stoken", Namespace: oneNS.Name}
			token := &v4alpha1.CloudflareServiceToken{
				ObjectMeta: metav1.ObjectMeta{
					Name:      sTokenNN.Name,
					Namespace: sTokenNN.Namespace,
				},
				Spec: v4alpha1.CloudflareServiceTokenSpec{
					Name: "ZTO AccessGroup Tests - 4 - Service Token",
				},
			}
			Expect(k8sClient.Create(ctx, token)).To(Not(HaveOccurred()))

			//
			ByExpectingCFResourceToBeReady(ctx,
				sTokenNN,
				token,
			).Should(Succeed())

			//
			By("Creating access group in namespace two, referencing sToken above")
			groupNN := types.NamespacedName{Name: "test-4-group", Namespace: twoNS.Name}
			group := &v4alpha1.CloudflareAccessGroup{
				ObjectMeta: metav1.ObjectMeta{
					Name:      groupNN.Name,
					Namespace: groupNN.Namespace,
				},
				Spec: v4alpha1.CloudflareAccessGroupSpec{
					Name: "ZTO AccessGroup Tests - 4 - Group",
					Include: v4alpha1.CloudFlareAccessRules{
						ServiceTokenRefs: []string{
							v4alpha1.ParsedNamespacedName(sTokenNN),
						},
					},
				},
			}
			Expect(k8sClient.Create(ctx, group)).To(Not(HaveOccurred()))

			//
			ByExpectingCFResourceToBeReady(ctx,
				groupNN,
				group,
			).Should(Succeed())
		})
	})
})
