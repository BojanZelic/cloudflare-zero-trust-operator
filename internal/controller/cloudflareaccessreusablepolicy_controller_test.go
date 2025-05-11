//go:build integration

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

var _ = Describe("CloudflareAccessReusablePolicy controller", Ordered, func() {
	//
	clear := func() {
		ctx := context.Background()

		By("Removing all existing access apps")
		apps, err := api.AccessApplications(ctx)
		Expect(err).To(Not(HaveOccurred()))
		for _, app := range *apps {
			err = api.DeleteAccessApplication(ctx, app.ID)
			// Expect(err).To(Not(HaveOccurred()))
		}

		By("Removing all existing access groups")
		groups, err := api.AccessGroups(ctx)
		Expect(err).To(Not(HaveOccurred()))
		for _, group := range *groups {
			_ = api.DeleteAccessGroup(ctx, group.ID)
			//Expect(err).To(Not(HaveOccurred()))
		}

		By("Removing all existing service tokens")
		tokens, err := api.ServiceTokens(ctx)
		Expect(err).To(Not(HaveOccurred()))
		for _, token := range *tokens {
			_ = api.DeleteAccessServiceToken(ctx, token.ID)
			//Expect(err).To(Not(HaveOccurred()))
		}
	}
	BeforeAll(clear)
	AfterAll(clear)

	//
	//
	//

	Context("CloudflareAccessReusablePolicy controller test", func() {

		const cloudflareName = "cloudflare-rp"

		ctx := context.Background()

		namespace := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name:      cloudflareName,
				Namespace: cloudflareName,
			},
		}

		BeforeEach(func() {
			logOutput.Clear()

			By("Creating the Namespace to perform the tests")
			k8sClient.Create(ctx, namespace)
			// ignore error because of https://book.kubebuilder.io/reference/envtest.html#namespace-usage-limitation
			//Expect(err).To(Not(HaveOccurred()))
		})

		AfterEach(func() {
			By("expect no reconcile errors occured")
			Expect(logOutput.GetErrorCount()).To(Equal(0), logOutput.GetOutput())
		})

		// AfterEach(func() {
		// 	By("Deleting the Namespace to perform the tests")
		// 	//_ = k8sClient.Delete(ctx, namespace)
		// })

		It("should fail to reconcile CloudflareAccessReusablePolicy policies with bad references", func() {
			typeNamespaceName := types.NamespacedName{Name: "cloudflare-rp-four", Namespace: cloudflareName}

			By("Creating the custom resource for the Kind CloudflareAccessReusablePolicy")
			apps := &v4alpha1.CloudflareAccessReusablePolicy{
				ObjectMeta: metav1.ObjectMeta{
					Name:      typeNamespaceName.Name,
					Namespace: namespace.Name,
				},
				Spec: v4alpha1.CloudflareAccessReusablePolicySpec{
					Name: "bad-reference policies",
					Include: v4alpha1.CloudFlareAccessRules{
						AccessGroups: []v4alpha1.AccessGroup{
							{
								ValueFrom: &v4alpha1.AccessGroupReference{
									Name:      "idontexist",
									Namespace: "inanynamespace",
								},
							},
						},
					},
				},
			}
			err := k8sClient.Create(ctx, apps)
			Expect(err).To(Not(HaveOccurred()))

			By("Checking the Status")
			Eventually(func(g Gomega) {
				err = k8sClient.Get(ctx, typeNamespaceName, apps)
				g.Expect(err).To(Not(HaveOccurred()))
				g.Expect(apps.Status.Conditions).ToNot(BeEmpty())
				g.Expect(apps.Status.Conditions[len(apps.Status.Conditions)-1].Status).To(Equal(metav1.ConditionFalse))
			}, time.Second*10, time.Second).Should(Succeed())
		})

		It("should successfully reconcile CloudflareAccessReusablePolicy policies with references", func() {
			By("pre-create an access group")
			group := &v4alpha1.CloudflareAccessGroup{}
			typeNamespaceName := types.NamespacedName{Name: "cloudflare-rp-three", Namespace: cloudflareName}

			group = &v4alpha1.CloudflareAccessGroup{
				ObjectMeta: metav1.ObjectMeta{
					Name:      typeNamespaceName.Name,
					Namespace: typeNamespaceName.Namespace,
				},
				Spec: v4alpha1.CloudflareAccessGroupSpec{
					Name: "reference test",
					Include: v4alpha1.CloudFlareAccessRules{
						Emails: []string{"test2@cf-operator-tests.uk"},
					},
				},
			}

			err := k8sClient.Create(ctx, group)
			Expect(err).To(Not(HaveOccurred()))

			err = k8sClient.Get(ctx, typeNamespaceName, group)
			Expect(err).To(Not(HaveOccurred()))

			//
			//
			//

			By("pre-create a service token")
			token := &v4alpha1.CloudflareServiceToken{
				ObjectMeta: metav1.ObjectMeta{
					Name:      typeNamespaceName.Name,
					Namespace: typeNamespaceName.Namespace,
				},
				Spec: v4alpha1.CloudflareServiceTokenSpec{
					Name: "reference test",
				},
			}

			err = k8sClient.Create(ctx, token)
			Expect(err).To(Not(HaveOccurred()))

			err = k8sClient.Get(ctx, typeNamespaceName, token)
			Expect(err).To(Not(HaveOccurred()))

			//
			//
			//

			By("Creating the custom resource for the Kind CloudflareAccessReusablePolicy")
			RPTypeNamespaceName := types.NamespacedName{Name: "reference_test_manifest", Namespace: cloudflareName}

			reusablePolicy := &v4alpha1.CloudflareAccessReusablePolicy{
				ObjectMeta: metav1.ObjectMeta{
					Name:      RPTypeNamespaceName.Name,
					Namespace: namespace.Name,
				},
				Spec: v4alpha1.CloudflareAccessReusablePolicySpec{
					Name:     "reference_test",
					Decision: "allow",
					Include: v4alpha1.CloudFlareAccessRules{
						AccessGroups: []v4alpha1.AccessGroup{
							{
								ValueFrom: &v4alpha1.AccessGroupReference{
									Name:      group.Name,
									Namespace: group.Namespace,
								},
							},
						},
						ServiceTokens: []v4alpha1.ServiceToken{
							{
								ValueFrom: &v4alpha1.ServiceTokenReference{
									Name:      token.Name,
									Namespace: token.Namespace,
								},
							},
						},
					},
				},
			}

			err = k8sClient.Create(ctx, reusablePolicy)
			Expect(err).To(Not(HaveOccurred()))

			RPFound := &v4alpha1.CloudflareAccessReusablePolicy{}
			By("Checking the latest Status should have the ID of the resource")
			Eventually(func() string {
				RPFound = &v4alpha1.CloudflareAccessReusablePolicy{}
				k8sClient.Get(ctx, RPTypeNamespaceName, RPFound)
				return RPFound.Status.AccessReusablePolicyID
			}, time.Second*10, time.Second).Should(Not(BeEmpty()))

			//
			//
			//

			By("Creating the custom resource for the Kind CloudflareAccessReusablePolicy")
			apps := &v4alpha1.CloudflareAccessReusablePolicy{
				ObjectMeta: metav1.ObjectMeta{
					Name:      typeNamespaceName.Name,
					Namespace: namespace.Name,
				},
				Spec: v4alpha1.CloudflareAccessReusablePolicySpec{
					Name:   "reference policies",
					Domain: "reference-policies.cf-operator-tests.uk",
					PolicyRefs: []string{
						RPTypeNamespaceName.Name,
					},
				},
			}
			err = k8sClient.Create(ctx, apps)
			Expect(err).To(Not(HaveOccurred()))

			By("Checking the Status")
			Eventually(func(g Gomega) {
				err = k8sClient.Get(ctx, typeNamespaceName, apps)
				g.Expect(err).To(Not(HaveOccurred()))
				g.Expect(apps.Status.Conditions).ToNot(BeEmpty())
				g.Expect(apps.Status.Conditions[0].Status).To(Equal(metav1.ConditionTrue))
			}, time.Second*10, time.Second).Should(Succeed())
		})

		It("should successfully reconcile a custom resource for CloudflareAccessReusablePolicy", func() {
			By("Creating the custom resource for the Kind CloudflareAccessReusablePolicy")
			typeNamespaceName := types.NamespacedName{Name: "cloudflare-rp-five", Namespace: cloudflareName}

			apps := &v4alpha1.CloudflareAccessReusablePolicy{
				ObjectMeta: metav1.ObjectMeta{
					Name:      typeNamespaceName.Name,
					Namespace: typeNamespaceName.Namespace,
				},
				Spec: v4alpha1.CloudflareAccessReusablePolicySpec{
					Name:   "integration test",
					Domain: "integration.cf-operator-tests.uk",
				},
			}
			err := k8sClient.Create(ctx, apps)
			Expect(err).To(Not(HaveOccurred()))

			By("Checking if the custom resource was successfully created")
			Eventually(func() error {
				found := &v4alpha1.CloudflareAccessReusablePolicy{}
				return k8sClient.Get(ctx, typeNamespaceName, found)
			}, time.Second*10, time.Second).Should(Succeed())

			found := &v4alpha1.CloudflareAccessReusablePolicy{}
			By("Checking the latest Status should have the ID of the resource")
			Eventually(func(g Gomega) {
				found = &v4alpha1.CloudflareAccessReusablePolicy{}
				g.Expect(k8sClient.Get(ctx, typeNamespaceName, found)).To(Not(HaveOccurred()))
				g.Expect(found.Status.AccessReusablePolicyID).ToNot(BeEmpty())
				g.Expect(found.Status.CreatedAt.Time).To(Equal(found.Status.UpdatedAt.Time))
			}, time.Second*10, time.Second).Should(Succeed())

			By("Cloudflare resource should equal the spec")
			cfResource, err := api.AccessApplication(ctx, found.Status.AccessReusablePolicyID)
			Expect(err).To(Not(HaveOccurred()))
			Expect(cfResource.Name).To(Equal(found.Spec.Name))

			By("Get the latest version of the resource")
			Expect(k8sClient.Get(ctx, typeNamespaceName, found)).To(Not(HaveOccurred()))
			By("Updating the name of the resource")
			found.Spec.Name = "updated name"
			Expect(k8sClient.Update(ctx, found)).To(Not(HaveOccurred()))

			By("Checking the latest Status should have the update")
			Eventually(func(g Gomega) {
				found = &v4alpha1.CloudflareAccessReusablePolicy{}
				g.Expect(k8sClient.Get(ctx, typeNamespaceName, found)).To(Not(HaveOccurred()))
				g.Expect(found.Spec.Name).To(Equal("updated name"))
				g.Expect(found.Status.AccessReusablePolicyID).ToNot(BeEmpty())
			}, time.Second*25, time.Second).Should(Succeed())

			By("Cloudflare resource should equal the updated spec")
			Eventually(func(g Gomega) {
				cfResource, err = api.AccessApplication(ctx, found.Status.AccessReusablePolicyID)
				g.Expect(err).To(Not(HaveOccurred()))
				g.Expect(cfResource.Name).To(Equal(found.Spec.Name))
			}, time.Second*45, time.Second).Should(Succeed(), logOutput.GetOutput()) //sometimes this is cached

			By("Cloudflare resource should be deleted")
			Expect(k8sClient.Delete(ctx, apps)).To(Not(HaveOccurred()))

			By("Checking if the custom resource was successfully deleted")
			Eventually(func() error {
				return k8sClient.Get(ctx, typeNamespaceName, apps)
			}, time.Second*10, time.Second).Should(Not(Succeed()))
		})

		It("should successfully reconcile CloudflareAccessReusablePolicy whose AccessReusablePolicyID references a missing Reusable Policy", func() {
			By("Recreating the custom resource for the Kind CloudflareAccessReusablePolicy")
			typeNamespaceName := types.NamespacedName{Name: "cloudflare-rp-seven", Namespace: cloudflareName}

			previousCreatedAndUpdatedDate := metav1.NewTime(time.Now().Add(-time.Hour * 24))
			apps := &v4alpha1.CloudflareAccessReusablePolicy{
				ObjectMeta: metav1.ObjectMeta{
					Name:      typeNamespaceName.Name,
					Namespace: namespace.Name,
				},
				Spec: v4alpha1.CloudflareAccessReusablePolicySpec{
					Name:    "missing application",
					Domains: []string{"recreate-application.cf-operator-tests.uk"},
				},
			}

			err := k8sClient.Create(ctx, apps)
			Expect(err).To(Not(HaveOccurred()))

			By("Checking if the custom resource was successfully created")
			Eventually(func() error {
				found := &v4alpha1.CloudflareAccessReusablePolicy{}
				return k8sClient.Get(ctx, typeNamespaceName, found)
			}, time.Minute, time.Second).Should(Succeed())

			found := &v4alpha1.CloudflareAccessReusablePolicy{}
			By("Checking the latest Status should have the ID of the resource")

			oldAccessReusablePolicyID := ""

			Eventually(func(g Gomega) {
				found = &v4alpha1.CloudflareAccessReusablePolicy{}
				g.Expect(k8sClient.Get(ctx, typeNamespaceName, found)).To(Not(HaveOccurred()))
				g.Expect(found.Status.AccessReusablePolicyID).ToNot(BeEmpty())
				oldAccessReusablePolicyID = found.Status.AccessReusablePolicyID
				g.Expect(found.Status.CreatedAt.Time).To(Equal(found.Status.UpdatedAt.Time))
				g.Expect(found.Status.CreatedAt.Time.After(previousCreatedAndUpdatedDate.Time)).To(BeTrue())
				g.Expect(found.Status.UpdatedAt.Time.After(previousCreatedAndUpdatedDate.Time)).To(BeTrue())
			}, time.Second*10, time.Second).Should(Succeed())

			Expect(api.DeleteAccessApplication(ctx, found.Status.AccessReusablePolicyID)).To(Not(HaveOccurred()))

			By("re-trigger reconcile by updating access application")
			Eventually(func(g Gomega) {
				found = &v4alpha1.CloudflareAccessReusablePolicy{}
				g.Expect(k8sClient.Get(ctx, typeNamespaceName, found)).To(Not(HaveOccurred()))
				found.Spec.Name = "updated name"
				Expect(k8sClient.Update(ctx, found)).To(Not(HaveOccurred()))
			}, time.Second*10, time.Second).Should(Succeed())

			By("Checking the latest Status should have the ID of the resource")
			Eventually(func(g Gomega) {
				found = &v4alpha1.CloudflareAccessReusablePolicy{}
				g.Expect(k8sClient.Get(ctx, typeNamespaceName, found)).To(Not(HaveOccurred()))
				g.Expect(found.Status.AccessReusablePolicyID).ToNot(Equal(oldAccessReusablePolicyID))
			}, time.Second*10, time.Second).Should(Succeed())

			By("Cloudflare resource should equal the updated spec")
			Eventually(func(g Gomega) {
				cfResource, err := api.AccessApplication(ctx, found.Status.AccessReusablePolicyID)
				g.Expect(err).To(Not(HaveOccurred()))
				g.Expect(cfResource.Name).To(Equal(found.Spec.Name))
			}, time.Second*45, time.Second).Should(Succeed(), logOutput.GetOutput()) //sometimes this is cached
		})
	})
})
