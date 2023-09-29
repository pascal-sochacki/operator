package controller

import (
	"context"
	"fmt"
	"os"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	cachev1alpha1 "github.com/example/memcached-operator/api/v1alpha1"
)

var _ = Describe("Memcached controller", func() {
	Context("Memcached controller test", func() {
		const MemcachedName = "test-memcached"
		ctx := context.Background()
		namespace := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name:      MemcachedName,
				Namespace: MemcachedName,
			},
		}

		typeNamespaceName := types.NamespacedName{Name: MemcachedName, Namespace: MemcachedName}
		BeforeEach(func() {
			By("Creating the Namespace to perform the tests")
			err := k8sClient.Create(ctx, namespace)
			Expect(err).To(Not(HaveOccurred()))

			By("Setting the Image ENV VAR which stores the Operand image")
			err = os.Setenv("MEMCACHED_IMAGE", "example.com/image:test")
			Expect(err).To(Not(HaveOccurred()))
		})

		It("should successfully reconcile a custom resource for Memcached", func() {
			By("Creating the custom resource for the Kind Memcached")
			memcached := &cachev1alpha1.Memcached{}
			err := k8sClient.Get(ctx, typeNamespaceName, memcached)
			if err != nil && errors.IsNotFound(err) {
				// Let's mock our custom resource at the same way that we would
				// apply on the cluster the manifest under config/samples
				memcached := &cachev1alpha1.Memcached{
					ObjectMeta: metav1.ObjectMeta{
						Name:      MemcachedName,
						Namespace: namespace.Name,
					},
					Spec: cachev1alpha1.MemcachedSpec{
						Size:          1,
						ContainerPort: 11211,
					},
				}

				err = k8sClient.Create(ctx, memcached)
				Expect(err).To(Not(HaveOccurred()))
			}
			By("Checking if the custom resource was successfully created")
			Eventually(func() error {
				found := &cachev1alpha1.Memcached{}
				return k8sClient.Get(ctx, typeNamespaceName, found)
			}, time.Minute, time.Second).Should(Succeed())

			By("Reconciling the custom resource created")
			memcachedReconciler := &MemcachedReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			_, err = memcachedReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespaceName,
			})
			Expect(err).To(Not(HaveOccurred()))

			By("Checking if Deployment was successfully created in the reconciliation")
			Eventually(func() error {
				found := &appsv1.Deployment{}
				return k8sClient.Get(ctx, typeNamespaceName, found)
			}, time.Minute, time.Second).Should(Succeed())

			By("Checking the latest Status Condition added to the Memcached instance")
			Eventually(func() error {
				_, err = memcachedReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: typeNamespaceName,
				})
				if err != nil {
					return fmt.Errorf("Could not reconcile")
				}
				err := k8sClient.Get(ctx, typeNamespaceName, memcached)
				if err != nil {
					return fmt.Errorf("Could not get memcached")
				}
				if memcached.Status.Conditions == nil {
					return fmt.Errorf("No Conditions found")
				}
				if len(memcached.Status.Conditions) != 0 {
					latestStatusCondition := memcached.Status.Conditions[0]
					expectedLatestStatusCondition := metav1.Condition{
						Type:    typeAvailableMemcached,
						Status:  metav1.ConditionTrue,
						Reason:  "Reconciling",
						Message: "Deployment for custom resource (test-memcached) with 1 replicas created successfully",
					}
					if latestStatusCondition.Message == expectedLatestStatusCondition.Message {
						return nil
					} else {
						return fmt.Errorf("The latest status condition added to the memcached instance is not as expected is: %s expected: %s", latestStatusCondition.Message, expectedLatestStatusCondition.Message)
					}
				}
				return fmt.Errorf("Conditions is empty")
			}, time.Minute, time.Second).Should(Succeed())

		})
	})
})
