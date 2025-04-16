//go:build integration

package controller

import (
	"context"
	"time"

	v1alpha1 "github.com/bojanzelic/cloudflare-zero-trust-operator/api/v1alpha1"
	"github.com/bojanzelic/cloudflare-zero-trust-operator/internal/cfcollections"
	cloudflare "github.com/cloudflare/cloudflare-go/v4"
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

	Context("CloudflareAccessApplication controller test", func() {

		const cloudflareName = "cloudflare-app"

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

			found := &v1alpha1.CloudflareAccessApplication{}
			By("Checking the latest Status should have the ID of the resource")
			Eventually(func() string {
				found = &v1alpha1.CloudflareAccessApplication{}
				k8sClient.Get(ctx, typeNamespaceName, found)
				return found.Status.AccessApplicationID
			}, time.Second*10, time.Second).Should(Not(BeEmpty()))

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
			}, time.Second*25, time.Second).Should(Succeed())

			By("changing a policy")
			k8sClient.Get(ctx, typeNamespaceName, found)
			found.Spec.Policies[0].Name = "updated_policy"
			found.Spec.Policies[0].Include[0].Emails[0] = "testemail3@cf-operator-tests.uk"
			err = k8sClient.Update(ctx, found)
			Expect(err).To(Not(HaveOccurred()))

			By("Cloudflare resource should equal the spec")
			Eventually(func(g Gomega) {
				cfResource, err = api.AccessPolicies(ctx, found.Status.AccessApplicationID)
				g.Expect(err).To(Not(HaveOccurred()))
				g.Expect(cfResource[0].Name).To(Equal(found.Spec.Policies[0].Name))
				g.Expect(cfResource[0].Include[0].(map[string]interface{})["email"].(map[string]interface{})["email"]).To(Equal(found.Spec.Policies[0].Include[0].Emails[0]))
				g.Expect(cfResource[0].Include[1].(map[string]interface{})["email"].(map[string]interface{})["email"]).To(Equal(found.Spec.Policies[0].Include[0].Emails[1]))
				g.Expect(cfResource[0].Include[2].(map[string]interface{})["email_domain"].(map[string]interface{})["domain"]).To(Equal(found.Spec.Policies[0].Include[1].EmailDomains[0]))
			}, time.Second*25, time.Second).Should(Succeed())
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
					Name:   "bad-reference policies",
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
			typeNamespaceName := types.NamespacedName{Name: "cloudflare-app-five", Namespace: cloudflareName}

			apps := &v1alpha1.CloudflareAccessApplication{
				ObjectMeta: metav1.ObjectMeta{
					Name:      typeNamespaceName.Name,
					Namespace: typeNamespaceName.Namespace,
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
			}, time.Second*10, time.Second).Should(Succeed())

			found := &v1alpha1.CloudflareAccessApplication{}
			By("Checking the latest Status should have the ID of the resource")
			Eventually(func(g Gomega) {
				found = &v1alpha1.CloudflareAccessApplication{}
				g.Expect(k8sClient.Get(ctx, typeNamespaceName, found)).To(Not(HaveOccurred()))
				g.Expect(found.Status.AccessApplicationID).ToNot(BeEmpty())
				g.Expect(found.Status.CreatedAt.Time).To(Equal(found.Status.UpdatedAt.Time))
			}, time.Second*10, time.Second).Should(Succeed())

			By("Cloudflare resource should equal the spec")
			cfResource, err := api.AccessApplication(ctx, found.Status.AccessApplicationID)
			Expect(err).To(Not(HaveOccurred()))
			Expect(cfResource.Name).To(Equal(found.Spec.Name))

			By("Get the latest version of the resource")
			Expect(k8sClient.Get(ctx, typeNamespaceName, found)).To(Not(HaveOccurred()))
			By("Updating the name of the resource")
			found.Spec.Name = "updated name"
			Expect(k8sClient.Update(ctx, found)).To(Not(HaveOccurred()))

			By("Checking the latest Status should have the update")
			Eventually(func(g Gomega) {
				found = &v1alpha1.CloudflareAccessApplication{}
				g.Expect(k8sClient.Get(ctx, typeNamespaceName, found)).To(Not(HaveOccurred()))
				g.Expect(found.Spec.Name).To(Equal("updated name"))
				g.Expect(found.Status.AccessApplicationID).ToNot(BeEmpty())
			}, time.Second*25, time.Second).Should(Succeed())

			By("Cloudflare resource should equal the updated spec")
			Eventually(func(g Gomega) {
				cfResource, err = api.AccessApplication(ctx, found.Status.AccessApplicationID)
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

		It("should successfully reconcile a custom resource for CloudflareAccessApplication", func() {
			By("Creating the custom resource for the Kind CloudflareAccessApplication")
			typeNamespaceName := types.NamespacedName{Name: "cloudflare-app-four", Namespace: cloudflareName}

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

			found := &v1alpha1.CloudflareAccessApplication{}
			By("Checking the latest Status should have the ID of the resource")
			Eventually(func(g Gomega) {
				found = &v1alpha1.CloudflareAccessApplication{}
				g.Expect(k8sClient.Get(ctx, typeNamespaceName, found)).To(Not(HaveOccurred()))
				g.Expect(found.Status.AccessApplicationID).ToNot(BeEmpty())
				g.Expect(found.Status.CreatedAt.Time).To(Equal(found.Status.UpdatedAt.Time))
			}, time.Second*10, time.Second).Should(Succeed())

			By("Cloudflare resource should equal the spec")
			cfResource, err := api.AccessApplication(ctx, found.Status.AccessApplicationID)
			Expect(err).To(Not(HaveOccurred()))
			Expect(cfResource.Name).To(Equal(found.Spec.Name))

			By("Get the latest version of the resource")
			Expect(k8sClient.Get(ctx, typeNamespaceName, found)).To(Not(HaveOccurred()))
			By("Updating the name of the resource")
			found.Spec.Name = "updated name"
			Expect(k8sClient.Update(ctx, found)).To(Not(HaveOccurred()))

			By("Checking the latest Status should have the update")
			Eventually(func(g Gomega) {
				found = &v1alpha1.CloudflareAccessApplication{}
				g.Expect(k8sClient.Get(ctx, typeNamespaceName, found)).To(Not(HaveOccurred()))
				g.Expect(found.Spec.Name).To(Equal("updated name"))
				g.Expect(found.Status.AccessApplicationID).ToNot(BeEmpty())
			}, time.Second*25, time.Second).Should(Succeed())

			By("Cloudflare resource should equal the updated spec")
			Eventually(func(g Gomega) {
				cfResource, err = api.AccessApplication(ctx, found.Status.AccessApplicationID)
				g.Expect(err).To(Not(HaveOccurred()))
				g.Expect(cfResource.Name).To(Equal(found.Spec.Name))
			}, time.Second*45, time.Second).Should(Succeed(), logOutput.GetOutput()) //sometimes this is cached
		})

		It("should be able to set a LogoURL for CloudflareAccessApplication", func() {
			By("Creating the custom resource for the Kind CloudflareAccessApplication")
			typeNamespaceName := types.NamespacedName{Name: "cloudflare-app-six", Namespace: cloudflareName}

			apps := &v1alpha1.CloudflareAccessApplication{
				ObjectMeta: metav1.ObjectMeta{
					Name:      typeNamespaceName.Name,
					Namespace: typeNamespaceName.Namespace,
				},
				Spec: v1alpha1.CloudflareAccessApplicationSpec{
					Name:    "integration test logo",
					Domain:  "integration-logo-test.cf-operator-tests.uk",
					LogoURL: "https://www.cloudflare.com/img/logo-web-badges/cf-logo-on-white-bg.svg",
				},
			}
			err := k8sClient.Create(ctx, apps)
			Expect(err).To(Not(HaveOccurred()))

			By("Checking if the custom resource was successfully created")
			Eventually(func() error {
				found := &v1alpha1.CloudflareAccessApplication{}
				return k8sClient.Get(ctx, typeNamespaceName, found)
			}, time.Minute, time.Second).Should(Succeed())

			found := &v1alpha1.CloudflareAccessApplication{}
			By("Checking the latest Status should have the ID of the resource")
			Eventually(func(g Gomega) {
				found = &v1alpha1.CloudflareAccessApplication{}
				g.Expect(k8sClient.Get(ctx, typeNamespaceName, found)).To(Not(HaveOccurred()))
				g.Expect(found.Status.AccessApplicationID).ToNot(BeEmpty())
				g.Expect(found.Status.CreatedAt.Time).To(Equal(found.Status.UpdatedAt.Time))
			}, time.Second*10, time.Second).Should(Succeed())

			By("Cloudflare resource should equal the spec")
			cfResource, err := api.AccessApplication(ctx, found.Status.AccessApplicationID)
			Expect(err).To(Not(HaveOccurred()))
			Expect(cfResource.Name).To(Equal(found.Spec.Name))
		})

		It("should successfully reconcile CloudflareAccessApplication whose AccessApplicationID references a missing Application", func() {
			By("Recreating the custom resource for the Kind CloudflareAccessApplication")
			typeNamespaceName := types.NamespacedName{Name: "cloudflare-app-seven", Namespace: cloudflareName}

			previousCreatedAndUpdatedDate := metav1.NewTime(time.Now().Add(-time.Hour * 24))
			apps := &v1alpha1.CloudflareAccessApplication{
				ObjectMeta: metav1.ObjectMeta{
					Name:      typeNamespaceName.Name,
					Namespace: namespace.Name,
				},
				Spec: v1alpha1.CloudflareAccessApplicationSpec{
					Name:   "missing application",
					Domain: "recreate-application.cf-operator-tests.uk",
				},
			}

			err := k8sClient.Create(ctx, apps)
			Expect(err).To(Not(HaveOccurred()))

			By("Checking if the custom resource was successfully created")
			Eventually(func() error {
				found := &v1alpha1.CloudflareAccessApplication{}
				return k8sClient.Get(ctx, typeNamespaceName, found)
			}, time.Minute, time.Second).Should(Succeed())

			found := &v1alpha1.CloudflareAccessApplication{}
			By("Checking the latest Status should have the ID of the resource")

			oldAccessApplicationID := ""

			Eventually(func(g Gomega) {
				found = &v1alpha1.CloudflareAccessApplication{}
				g.Expect(k8sClient.Get(ctx, typeNamespaceName, found)).To(Not(HaveOccurred()))
				g.Expect(found.Status.AccessApplicationID).ToNot(BeEmpty())
				oldAccessApplicationID = found.Status.AccessApplicationID
				g.Expect(found.Status.CreatedAt.Time).To(Equal(found.Status.UpdatedAt.Time))
				g.Expect(found.Status.CreatedAt.Time.After(previousCreatedAndUpdatedDate.Time)).To(BeTrue())
				g.Expect(found.Status.UpdatedAt.Time.After(previousCreatedAndUpdatedDate.Time)).To(BeTrue())
			}, time.Second*10, time.Second).Should(Succeed())

			Expect(api.DeleteAccessApplication(ctx, found.Status.AccessApplicationID)).To(Not(HaveOccurred()))

			By("re-trigger reconcile by updating access application")
			Eventually(func(g Gomega) {
				found = &v1alpha1.CloudflareAccessApplication{}
				g.Expect(k8sClient.Get(ctx, typeNamespaceName, found)).To(Not(HaveOccurred()))
				found.Spec.Name = "updated name"
				Expect(k8sClient.Update(ctx, found)).To(Not(HaveOccurred()))
			}, time.Second*10, time.Second).Should(Succeed())

			By("Checking the latest Status should have the ID of the resource")
			Eventually(func(g Gomega) {
				found = &v1alpha1.CloudflareAccessApplication{}
				g.Expect(k8sClient.Get(ctx, typeNamespaceName, found)).To(Not(HaveOccurred()))
				g.Expect(found.Status.AccessApplicationID).ToNot(Equal(oldAccessApplicationID))
			}, time.Second*10, time.Second).Should(Succeed())

			By("Cloudflare resource should equal the updated spec")
			Eventually(func(g Gomega) {
				cfResource, err := api.AccessApplication(ctx, found.Status.AccessApplicationID)
				g.Expect(err).To(Not(HaveOccurred()))
				g.Expect(cfResource.Name).To(Equal(found.Spec.Name))
			}, time.Second*45, time.Second).Should(Succeed(), logOutput.GetOutput()) //sometimes this is cached
		})
	})
})
