package controllers

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client"

	samplecontrollerv1alpha1 "github.com/govargo/sample-controller-kubebuilder/api/v1alpha1"
)

var _ = Context("Inside of a new namespace", func() {
	ctx := context.TODO()
	ns := SetupTest(ctx)

	Describe("when no existing resources exist", func() {

		It("should create a new Deployment resource with the specified name and one replica if one replica is provided", func() {
			foo := &samplecontrollerv1alpha1.Foo{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "example-foo",
					Namespace: ns.Name,
				},
				Spec: samplecontrollerv1alpha1.FooSpec{
					DeploymentName: "example-foo",
					Replicas:       pointer.Int32Ptr(1),
				},
			}

			err := k8sClient.Create(ctx, foo)
			Expect(err).NotTo(HaveOccurred(), "failed to create test Foo resource")

			deployment := &appsv1.Deployment{}
			Eventually(
				getResourceFunc(ctx, client.ObjectKey{Name: "example-foo", Namespace: foo.Namespace}, deployment),
				time.Second*5, time.Millisecond*500).Should(BeNil())

			Expect(*deployment.Spec.Replicas).To(Equal(int32(1)))
		})

		It("should create a new Deployment resource with the specified name and two replicas if two is specified", func() {
			foo := &samplecontrollerv1alpha1.Foo{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "example-foo",
					Namespace: ns.Name,
				},
				Spec: samplecontrollerv1alpha1.FooSpec{
					DeploymentName: "example-foo",
					Replicas:       pointer.Int32Ptr(2),
				},
			}

			err := k8sClient.Create(ctx, foo)
			Expect(err).NotTo(HaveOccurred(), "failed to create test Foo resource")

			deployment := &appsv1.Deployment{}
			Eventually(
				getResourceFunc(ctx, client.ObjectKey{Name: "example-foo", Namespace: foo.Namespace}, deployment),
				time.Second*5, time.Millisecond*500).Should(BeNil())

			Expect(*deployment.Spec.Replicas).To(Equal(int32(2)))
		})

		It("should allow updating the replicas count after creating a Foo resource", func() {
			deploymentObjectKey := client.ObjectKey{
				Name:      "example-foo",
				Namespace: ns.Name,
			}
			fooObjectKey := client.ObjectKey{
				Name:      "example-foo",
				Namespace: ns.Name,
			}
			foo := &samplecontrollerv1alpha1.Foo{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fooObjectKey.Name,
					Namespace: fooObjectKey.Namespace,
				},
				Spec: samplecontrollerv1alpha1.FooSpec{
					DeploymentName: deploymentObjectKey.Name,
					Replicas:       pointer.Int32Ptr(1),
				},
			}

			err := k8sClient.Create(ctx, foo)
			Expect(err).NotTo(HaveOccurred(), "failed to create test Foo resource")

			deployment := &appsv1.Deployment{}
			Eventually(
				getResourceFunc(ctx, deploymentObjectKey, deployment),
				time.Second*5, time.Millisecond*500).Should(BeNil(), "deployment resource should exist")

			Expect(*deployment.Spec.Replicas).To(Equal(int32(1)), "replica count should be equal to 1")

			err = k8sClient.Get(ctx, fooObjectKey, foo)
			Expect(err).NotTo(HaveOccurred(), "failed to retrieve Foo resource")

			foo.Spec.Replicas = pointer.Int32Ptr(2)
			err = k8sClient.Update(ctx, foo)
			Expect(err).NotTo(HaveOccurred(), "failed to Update Foo resource")

			Eventually(getDeploymentReplicasFunc(ctx, deploymentObjectKey)).
				Should(Equal(int32(2)), "expected Deployment resource to be scale to 2 replicas")
		})

		It("should clean up an old Deployment resource if the deploymentName is changed", func() {
			deploymentObjectKey := client.ObjectKey{
				Name:      "example-foo",
				Namespace: ns.Name,
			}
			newDeploymentObjectKey := client.ObjectKey{
				Name:      "example-foo-new",
				Namespace: ns.Name,
			}
			fooObjectKey := client.ObjectKey{
				Name:      "example-foo",
				Namespace: ns.Name,
			}
			foo := &samplecontrollerv1alpha1.Foo{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fooObjectKey.Name,
					Namespace: fooObjectKey.Namespace,
				},
				Spec: samplecontrollerv1alpha1.FooSpec{
					DeploymentName: deploymentObjectKey.Name,
					Replicas:       pointer.Int32Ptr(1),
				},
			}

			err := k8sClient.Create(ctx, foo)
			Expect(err).NotTo(HaveOccurred(), "failed to create test Foo resource")

			deployment := &appsv1.Deployment{}
			Eventually(
				getResourceFunc(ctx, deploymentObjectKey, deployment),
				time.Second*5, time.Millisecond*500).Should(BeNil(), "deployment resource should exist")

			err = k8sClient.Get(ctx, fooObjectKey, foo)
			Expect(err).NotTo(HaveOccurred(), "failed to retrieve Foo resource")

			foo.Spec.DeploymentName = newDeploymentObjectKey.Name
			err = k8sClient.Update(ctx, foo)
			Expect(err).NotTo(HaveOccurred(), "failed to Update Foo resource")

			Eventually(
				getResourceFunc(ctx, deploymentObjectKey, deployment),
				time.Second*5, time.Millisecond*500).ShouldNot(BeNil(), "old deployment resource should be deleted")

			Eventually(
				getResourceFunc(ctx, newDeploymentObjectKey, deployment),
				time.Second*5, time.Millisecond*500).Should(BeNil(), "new deployment resource should be created")
		})
	})
})

func getResourceFunc(ctx context.Context, key client.ObjectKey, obj runtime.Object) func() error {
	return func() error {
		return k8sClient.Get(ctx, key, obj)
	}
}

func getDeploymentReplicasFunc(ctx context.Context, key client.ObjectKey) func() int32 {
	return func() int32 {
		depl := &appsv1.Deployment{}
		err := k8sClient.Get(ctx, key, depl)
		Expect(err).NotTo(HaveOccurred(), "failed to get Deployment resource")

		return *depl.Spec.Replicas
	}
}
