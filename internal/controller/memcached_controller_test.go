package controller

import (
	"context"
	"os"

	corev1 "k8s.io/api/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
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

		//	typeNamespaceName := types.NamespacedName{Name: MemcachedName, Namespace: MemcachedName}
		BeforeEach(func() {
			By("Creating the Namespace to perform the tests")
			err := k8sClient.Create(ctx, namespace)
			Expect(err).To(Not(HaveOccurred()))

			By("Setting the Image ENV VAR which stores the Operand image")
			err = os.Setenv("MEMCACHED_IMAGE", "example.com/image:test")
			Expect(err).To(Not(HaveOccurred()))
		})

		It("should successfully reconcile", func() {
		})
	})
})
