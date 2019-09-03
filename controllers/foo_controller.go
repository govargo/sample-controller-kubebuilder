/*

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	samplecontrollerv1alpha1 "github.com/govargo/sample-controller-kubebuilder/api/v1alpha1"
)

// FooReconciler reconciles a Foo object
type FooReconciler struct {
	client.Client
	Log      logr.Logger
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

// +kubebuilder:rbac:groups=samplecontroller.k8s.io,resources=foos,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=samplecontroller.k8s.io,resources=foos/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;delete
// +kubebuilder:rbac:groups="",resources=events,verbs=create;patch

func (r *FooReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("foo", req.NamespacedName)

	/*
		### 1: Load the Foo by name
		We'll fetch the Foo using our client.
		All client methods take a context (to allow for cancellation) as
		their first argument, and the object
		in question as their last.
		Get is a bit special, in that it takes a
		[`NamespacedName`](https://godoc.org/sigs.k8s.io/controller-runtime/pkg/client#ObjectKey)
		as the middle argument (most don't have a middle argument, as we'll see below).
		Many client methods also take variadic options at the end.
	*/
	var foo samplecontrollerv1alpha1.Foo
	log.Info("fetching Foo Resource")
	if err := r.Get(ctx, req.NamespacedName, &foo); err != nil {
		log.Error(err, "unable to fetch Foo")
		// we'll ignore not-found errors, since they can't be fixed by an immediate
		// requeue (we'll need to wait for a new notification), and we can get them
		// on deleted requests.
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	/*
		### 2: Clean Up old Deployment which had been owned by Foo Resource.
		We'll find deployment object which foo object owns.
		If there is a deployment which is owned by foo and it doesn't match foo.spec.deploymentName,
		we clean up the deployment object.
		(If we do nothing without this func, the old deployment object keeps existing.)
	*/
	if err := r.cleanupOwnedResources(ctx, log, &foo); err != nil {
		log.Error(err, "failed to clean up old Deployment resources for this Foo")
		return ctrl.Result{}, err
	}

	/*
		### 3: Create or Update deployment object which match foo.Spec.
		We'll use ctrl.CreateOrUpdate method.
		It enable us to create an object if it doesn't exist.
		Or it enable us to update the object if it exists.
	*/

	// get deploymentName from foo.Spec
	deploymentName := foo.Spec.DeploymentName

	// define deployment template using deploymentName
	deploy := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      deploymentName,
			Namespace: req.Namespace,
		},
	}

	// Create or Update deployment object
	if _, err := ctrl.CreateOrUpdate(ctx, r.Client, deploy, func() error {

		// set the replicas from foo.Spec
		replicas := int32(1)
		if foo.Spec.Replicas != nil {
			replicas = *foo.Spec.Replicas
		}
		deploy.Spec.Replicas = &replicas

		// set a label for our deployment
		labels := map[string]string{
			"app":        "nginx",
			"controller": req.Name,
		}

		// set labels to spec.selector for our deployment
		if deploy.Spec.Selector == nil {
			deploy.Spec.Selector = &metav1.LabelSelector{MatchLabels: labels}
		}

		// set labels to template.objectMeta for our deployment
		if deploy.Spec.Template.ObjectMeta.Labels == nil {
			deploy.Spec.Template.ObjectMeta.Labels = labels
		}

		// set a container for our deployment
		containers := []corev1.Container{
			{
				Name:  "nginx",
				Image: "nginx:latest",
			},
		}

		// set containers to template.spec.containers for our deployment
		if deploy.Spec.Template.Spec.Containers == nil {
			deploy.Spec.Template.Spec.Containers = containers
		}

		// set the owner so that garbage collection can kicks in
		if err := ctrl.SetControllerReference(&foo, deploy, r.Scheme); err != nil {
			log.Error(err, "unable to set ownerReference from Foo to Deployment")
			return err
		}

		// end of ctrl.CreateOrUpdate
		return nil

	}); err != nil {

		// error handling of ctrl.CreateOrUpdate
		log.Error(err, "unable to ensure deployment is correct")
		return ctrl.Result{}, err

	}

	/*
		### 4: Update foo status.
		First, we get deployment object from in-memory-cache.
		Second, we get deployment.status.AvailableReplicas in order to update foo.status.AvailableReplicas.
		Third, we update foo.status from deployment.status.AvailableReplicas.
		Finally, finish reconcile. and the next reconcile loop would start unless controller process ends.
	*/

	// get deployment object from in-memory-cache
	var deployment appsv1.Deployment
	var deploymentNamespacedName = client.ObjectKey{Namespace: req.Namespace, Name: foo.Spec.DeploymentName}
	if err := r.Get(ctx, deploymentNamespacedName, &deployment); err != nil {
		log.Error(err, "unable to fetch Deployment")
		// we'll ignore not-found errors, since they can't be fixed by an immediate
		// requeue (we'll need to wait for a new notification), and we can get them
		// on deleted requests.
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// set foo.status.AvailableReplicas from deployment
	availableReplicas := deployment.Status.AvailableReplicas
	if availableReplicas == foo.Status.AvailableReplicas {
		// if availableReplicas equals availableReplicas, we wouldn't update anything.
		// exit Reconcile func without updating foo.status
		return ctrl.Result{}, nil
	}
	foo.Status.AvailableReplicas = availableReplicas

	// update foo.status
	if err := r.Status().Update(ctx, &foo); err != nil {
		log.Error(err, "unable to update Foo status")
		return ctrl.Result{}, err
	}

	// create event for updated foo.status
	r.Recorder.Eventf(&foo, corev1.EventTypeNormal, "Updated", "Update foo.status.AvailableReplicas: %d", foo.Status.AvailableReplicas)

	return ctrl.Result{}, nil
}

// cleanupOwnedResources will delete any existing Deployment resources that
// were created for the given Foo that no longer match the
// foo.spec.deploymentName field.
func (r *FooReconciler) cleanupOwnedResources(ctx context.Context, log logr.Logger, foo *samplecontrollerv1alpha1.Foo) error {
	log.Info("finding existing Deployments for Foo resource")

	// List all deployment resources owned by this Foo
	var deployments appsv1.DeploymentList
	if err := r.List(ctx, &deployments, client.InNamespace(foo.Namespace), client.MatchingFields(map[string]string{deploymentOwnerKey: foo.Name})); err != nil {
		return err
	}

	// Delete deployment if the deployment name doesn't match foo.spec.deploymentName
	for _, deployment := range deployments.Items {
		if deployment.Name == foo.Spec.DeploymentName {
			// If this deployment's name matches the one on the Foo resource
			// then do not delete it.
			continue
		}

		// Delete old deployment object which doesn't match foo.spec.deploymentName
		if err := r.Delete(ctx, &deployment); err != nil {
			log.Error(err, "failed to delete Deployment resource")
			return err
		}

		log.Info("delete deployment resource: " + deployment.Name)
		r.Recorder.Eventf(foo, corev1.EventTypeNormal, "Deleted", "Deleted deployment %q", deployment.Name)
	}

	return nil
}

var (
	deploymentOwnerKey = ".metadata.controller"
	apiGVStr           = samplecontrollerv1alpha1.GroupVersion.String()
)

// setup with controller manager
func (r *FooReconciler) SetupWithManager(mgr ctrl.Manager) error {

	// add deploymentOwnerKey index to deployment object which foo resource owns
	if err := mgr.GetFieldIndexer().IndexField(&appsv1.Deployment{}, deploymentOwnerKey, func(rawObj runtime.Object) []string {
		// grab the deployment object, extract the owner...
		deployment := rawObj.(*appsv1.Deployment)
		owner := metav1.GetControllerOf(deployment)
		if owner == nil {
			return nil
		}
		// ...make sure it's a Foo...
		if owner.APIVersion != apiGVStr || owner.Kind != "Foo" {
			return nil
		}

		// ...and if so, return it
		return []string{owner.Name}
	}); err != nil {
		return err
	}

	// define to watch targets...Foo resource and owned Deployment
	return ctrl.NewControllerManagedBy(mgr).
		For(&samplecontrollerv1alpha1.Foo{}).
		Owns(&appsv1.Deployment{}).
		Complete(r)
}
