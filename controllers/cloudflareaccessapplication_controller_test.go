//go:build integration

package controllers

import (
	"context"
	"time"

	v1alpha1 "github.com/bojanzelic/cloudflare-zero-trust-operator/api/v1alpha1"
	"github.com/bojanzelic/cloudflare-zero-trust-operator/internal/cfcollections"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
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

	AfterAll(func() {
		ctx := context.Background()

		By("Removing all existing access apps")
		apps, err := api.AccessApplications(ctx)
		Expect(err).To(Not(HaveOccurred()))
		for _, app := range apps {
			err = api.DeleteAccessApplication(ctx, app.ID)
			Expect(err).To(Not(HaveOccurred()))
		}
	})

	// BeforeEach(func() {
	// 	logOutput.Clear()
	// })

	// AfterEach(func() {
	// 	fmt.Println("called")
	// 	Expect(logOutput.GetErrorCount()).To(Equal(0))
	// })

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

		It("should successfully reconcile CloudflareAccessApplication policies", func() {
			By("Creating the custom resource for the Kind CloudflareAccessApplication")
			typeNamespaceName := types.NamespacedName{Name: "cloudflare-app-two", Namespace: cloudflareName}
			apps := &v1alpha1.CloudflareAccessApplication{
				ObjectMeta: metav1.ObjectMeta{
					Name:      typeNamespaceName.Name,
					Namespace: namespace.Name,
				},
				Spec: v1alpha1.CloudflareAccessApplicationSpec{
					Name:   "integration policies test",
					Domain: "integration-policies.cf-operator-tests.uk",
					Policies: v1alpha1.CloudflareAccessPolicyList{
						{
							Name:     "integration_test",
							Decision: "allow",
							Include: []v1alpha1.CloudFlareAccessGroupRule{
								{
									Emails: []string{"testemail@cf-operator-tests.uk", "testemail2@cf-operator-tests.uk"},
								},
								{
									EmailDomains: []string{"cf-operator-tests.uk"},
								},
							},
						},
					},
				},
			}
			err := k8sClient.Create(ctx, apps)
			Expect(err).To(Not(HaveOccurred()))

			// By("Reconciling the custom resource created")
			// accessAppReconciler := &CloudflareAccessApplicationReconciler{
			// 	Client: k8sClient,
			// 	Scheme: k8sClient.Scheme(),
			// }

			// _, err = accessAppReconciler.Reconcile(ctx, reconcile.Request{
			// 	NamespacedName: typeNamespaceName,
			// })
			// Expect(err).To(Not(HaveOccurred()))

			found := &v1alpha1.CloudflareAccessApplication{}
			By("Checking the latest Status should have the ID of the resource")
			Eventually(func() string {
				found = &v1alpha1.CloudflareAccessApplication{}
				k8sClient.Get(ctx, typeNamespaceName, found)
				return found.Status.AccessApplicationID
			}, time.Minute, time.Second).Should(Not(BeEmpty()))

			var cfResource cfcollections.AccessPolicyCollection
			By("Cloudflare resource should equal the spec")
			Eventually(func(g Gomega) {
				cfResource, err = api.AccessPolicies(ctx, found.Status.AccessApplicationID)
				g.Expect(err).To(Not(HaveOccurred()))
				g.Expect(cfResource).ToNot(BeEmpty())
				g.Expect(found.Spec.Policies).ToNot(BeEmpty())
				g.Expect(cfResource[0].Name).To(Equal(found.Spec.Policies[0].Name))
				g.Expect(cfResource[0].Include[0].(map[string]interface{})["email"].(map[string]interface{})["email"]).To(Equal(found.Spec.Policies[0].Include[0].Emails[0]))
				g.Expect(cfResource[0].Include[1].(map[string]interface{})["email"].(map[string]interface{})["email"]).To(Equal(found.Spec.Policies[0].Include[0].Emails[1]))
				g.Expect(cfResource[0].Include[2].(map[string]interface{})["email_domain"].(map[string]interface{})["domain"]).To(Equal(found.Spec.Policies[0].Include[1].EmailDomains[0]))
			}).Should(Succeed())

			By("changing a policy")
			k8sClient.Get(ctx, typeNamespaceName, found)
			found.Spec.Policies[0].Name = "updated_policy"
			found.Spec.Policies[0].Include[0].Emails[0] = "testemail3@cf-operator-tests.uk"
			err = k8sClient.Update(ctx, found)
			Expect(err).To(Not(HaveOccurred()))

			// By("Reconciling the updated custom resource")
			// _, err = accessAppReconciler.Reconcile(ctx, reconcile.Request{
			// 	NamespacedName: typeNamespaceName,
			// })
			// Expect(err).To(Not(HaveOccurred()))

			By("Cloudflare resource should equal the spec")
			Eventually(func(g Gomega) {
				cfResource, err = api.AccessPolicies(ctx, found.Status.AccessApplicationID)
				g.Expect(err).To(Not(HaveOccurred()))
				g.Expect(cfResource[0].Name).To(Equal(found.Spec.Policies[0].Name))
				g.Expect(cfResource[0].Include[0].(map[string]interface{})["email"].(map[string]interface{})["email"]).To(Equal(found.Spec.Policies[0].Include[0].Emails[0]))
				g.Expect(cfResource[0].Include[1].(map[string]interface{})["email"].(map[string]interface{})["email"]).To(Equal(found.Spec.Policies[0].Include[0].Emails[1]))
				g.Expect(cfResource[0].Include[2].(map[string]interface{})["email_domain"].(map[string]interface{})["domain"]).To(Equal(found.Spec.Policies[0].Include[1].EmailDomains[0]))
			}, time.Second*10, time.Second).Should(Succeed())
		})

		It("should fail to reconcile CloudflareAccessApplication policies with bad references", func() {
			typeNamespaceName := types.NamespacedName{Name: "cloudflare-app-four", Namespace: cloudflareName}

			By("Creating the custom resource for the Kind CloudflareAccessApplication")
			apps := &v1alpha1.CloudflareAccessApplication{
				ObjectMeta: metav1.ObjectMeta{
					Name:      typeNamespaceName.Name,
					Namespace: namespace.Name,
				},
				Spec: v1alpha1.CloudflareAccessApplicationSpec{
					Name:   "bad-reference policies ",
					Domain: "bad-reference-policies.cf-operator-tests.uk",
					Policies: v1alpha1.CloudflareAccessPolicyList{
						{
							Name:     "reference_test",
							Decision: "allow",
							Include: []v1alpha1.CloudFlareAccessGroupRule{{
								AccessGroups: []v1alpha1.AccessGroup{
									{
										ValueFrom: &v1alpha1.AccessGroupReference{
											Name:      "idontexist",
											Namespace: "inanynamespace",
										},
									},
								},
							}},
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

		It("should successfully reconcile CloudflareAccessApplication policies with references", func() {

			By("pre-create an access group")
			group := &v1alpha1.CloudflareAccessGroup{}
			typeNamespaceName := types.NamespacedName{Name: "cloudflare-app-three", Namespace: cloudflareName}

			group = &v1alpha1.CloudflareAccessGroup{
				ObjectMeta: metav1.ObjectMeta{
					Name:      typeNamespaceName.Name,
					Namespace: typeNamespaceName.Namespace,
				},
				Spec: v1alpha1.CloudflareAccessGroupSpec{
					Name: "reference test",
					Include: []v1alpha1.CloudFlareAccessGroupRule{
						{
							Emails: []string{"test2@cf-operator-tests.uk"},
						},
					},
				},
			}

			err := k8sClient.Create(ctx, group)
			Expect(err).To(Not(HaveOccurred()))

			// By("Reconciling the custom resource created")
			// accessGroupReconciler := &CloudflareAccessGroupReconciler{
			// 	Client: k8sClient,
			// 	Scheme: k8sClient.Scheme(),
			// }

			// _, err = accessGroupReconciler.Reconcile(ctx, reconcile.Request{
			// 	NamespacedName: typeNamespaceName,
			// })
			// Expect(err).To(Not(HaveOccurred()))

			err = k8sClient.Get(ctx, typeNamespaceName, group)
			Expect(err).To(Not(HaveOccurred()))

			By("pre-create a service token")
			token := &v1alpha1.CloudflareServiceToken{
				ObjectMeta: metav1.ObjectMeta{
					Name:      typeNamespaceName.Name,
					Namespace: typeNamespaceName.Namespace,
				},
				Spec: v1alpha1.CloudflareServiceTokenSpec{
					Name: "reference test",
				},
			}

			err = k8sClient.Create(ctx, token)
			Expect(err).To(Not(HaveOccurred()))

			// By("Reconciling the custom resource created")
			// serviceTokenReconciler := &CloudflareServiceTokenReconciler{
			// 	Client: k8sClient,
			// 	Scheme: k8sClient.Scheme(),
			// }

			// _, err = serviceTokenReconciler.Reconcile(ctx, reconcile.Request{
			// 	NamespacedName: typeNamespaceName,
			// })
			// Expect(err).To(Not(HaveOccurred()))

			err = k8sClient.Get(ctx, typeNamespaceName, token)
			Expect(err).To(Not(HaveOccurred()))

			By("Creating the custom resource for the Kind CloudflareAccessApplication")
			apps := &v1alpha1.CloudflareAccessApplication{
				ObjectMeta: metav1.ObjectMeta{
					Name:      typeNamespaceName.Name,
					Namespace: namespace.Name,
				},
				Spec: v1alpha1.CloudflareAccessApplicationSpec{
					Name:   "reference policies ",
					Domain: "reference-policies.cf-operator-tests.uk",
					Policies: v1alpha1.CloudflareAccessPolicyList{
						{
							Name:     "reference_test",
							Decision: "allow",
							Include: []v1alpha1.CloudFlareAccessGroupRule{{
								AccessGroups: []v1alpha1.AccessGroup{
									{
										ValueFrom: &v1alpha1.AccessGroupReference{
											Name:      group.Name,
											Namespace: group.Namespace,
										},
									},
								},
								ServiceToken: []v1alpha1.ServiceToken{
									{
										ValueFrom: &v1alpha1.ServiceTokenReference{
											Name:      token.Name,
											Namespace: token.Namespace,
										},
									},
								},
							}},
						},
					},
				},
			}
			err = k8sClient.Create(ctx, apps)
			Expect(err).To(Not(HaveOccurred()))

			// By("Reconciling the custom resource created")
			// accessAppReconciler := &CloudflareAccessApplicationReconciler{
			// 	Client: k8sClient,
			// 	Scheme: k8sClient.Scheme(),
			// }

			// _, err = accessAppReconciler.Reconcile(ctx, reconcile.Request{
			// 	NamespacedName: typeNamespaceName,
			// })
			// Expect(err).To(Not(HaveOccurred()))

			By("Checking the Status")
			Eventually(func(g Gomega) {
				err = k8sClient.Get(ctx, typeNamespaceName, apps)
				g.Expect(err).To(Not(HaveOccurred()))
				g.Expect(apps.Status.Conditions).ToNot(BeEmpty())
				g.Expect(apps.Status.Conditions[0].Status).To(Equal(metav1.ConditionTrue))
			}, time.Second*10, time.Second).Should(Succeed())
		})

		It("should successfully reconcile a custom resource for CloudflareAccessApplication", func() {
			By("Creating the custom resource for the Kind CloudflareAccessApplication")
			apps := &v1alpha1.CloudflareAccessApplication{
				ObjectMeta: metav1.ObjectMeta{
					Name:      cloudflareName,
					Namespace: namespace.Name,
				},
				Spec: v1alpha1.CloudflareAccessApplicationSpec{
					Name:   "integration test",
					Domain: "integration.cf-operator-tests.uk",
				},
			}
			err := k8sClient.Create(ctx, apps)
			Expect(err).To(Not(HaveOccurred()))

			By("Checking if the custom resource was successfully created")
			Eventually(func() error {
				found := &v1alpha1.CloudflareAccessApplication{}
				return k8sClient.Get(ctx, typeNamespaceName, found)
			}, time.Minute, time.Second).Should(Succeed())

			// By("Reconciling the custom resource created")
			// accessGroupReconciler := &CloudflareAccessApplicationReconciler{
			// 	Client: k8sClient,
			// 	Scheme: k8sClient.Scheme(),
			// }

			// _, err = accessGroupReconciler.Reconcile(ctx, reconcile.Request{
			// 	NamespacedName: typeNamespaceName,
			// })
			// Expect(err).To(Not(HaveOccurred()))

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
			Expect(cfResource.Name).To(Equal(found.Spec.Name))

			By("Updating the name of the resource")
			found.Spec.Name = "updated name"
			k8sClient.Update(ctx, found)
			Expect(err).To(Not(HaveOccurred()))

			// By("Reconciling the updated resource")
			// _, err = accessGroupReconciler.Reconcile(ctx, reconcile.Request{
			// 	NamespacedName: typeNamespaceName,
			// })
			// Expect(err).To(Not(HaveOccurred()))

			By("Cloudflare resource should equal the updated spec")
			Eventually(func(g Gomega) {
				cfResource, err = api.AccessApplication(ctx, found.Status.AccessApplicationID)
				g.Expect(err).To(Not(HaveOccurred()))
				g.Expect(cfResource.Name).To(Equal(found.Spec.Name))
			}, time.Second*45, time.Second).Should(Succeed(), logOutput.GetOutput()) //sometimes this is cached
		})
	})
})
