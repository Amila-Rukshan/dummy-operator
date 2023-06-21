/*
Copyright 2023.

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
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	interviewv1alpha1 "github.com/Amila-Rukshan/dummy-operator/api/v1alpha1"
)

const dummyFinalizer = "interview.com/dummy-finalizer"

// DummyReconciler reconciles a Dummy object
type DummyReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=interview.com,resources=dummies,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=interview.com,resources=dummies/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=interview.com,resources=dummies/finalizers,verbs=update
//+kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Dummy object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.1/pkg/reconcile
func (r *DummyReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	// get the dummy object
	// if it is present in the cache thn=en log the message
	dummy := interviewv1alpha1.Dummy{}
	err := r.Get(ctx, req.NamespacedName, &dummy)
	if err != nil {
		if errors.IsNotFound(err) {
			// dummy was deleted or not created
			// stop the reconciliation
			log.Info("Dummy resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}

		log.Error(err, "Unable to fetch Dummy: %s", req.NamespacedName)
		return ctrl.Result{}, err
	}

	// after this point, dummpy object is present, it got created or updated

	// log the spec.Message
	log.Info(fmt.Sprintf("Dummy resource with name: %s in namespace: %s has the spec.Message: %s", req.Name, req.Namespace, dummy.Spec.Message))

	// check if the pod is already present
	pod := corev1.Pod{}
	err = r.Get(ctx, req.NamespacedName, &pod)
	if err != nil {
		if errors.IsNotFound(err) {
			// pod was deleted or not created
			// create the pod resoruce artifact
			pod, err := r.createNginxPod(&dummy)
			if err != nil {
				log.Error(err, "Failed to define new Pod resource for the Dummy object")
				return ctrl.Result{}, err
			}

			// create the pod
			if err := r.Create(ctx, pod); err != nil {
				log.Error(err, "Failed to create the pod")
				return ctrl.Result{}, err
			}

			// pod created successfully
			return ctrl.Result{}, nil
		}

		log.Error(err, "Unable to fetch Pod: %s", req.NamespacedName)
		return ctrl.Result{}, err
	}

	// if the pod does not include the dummy object as owner reference then add it,
	// to take control over the existing pod
	if !metav1.IsControlledBy(&pod, &dummy) {
		if err := controllerutil.SetControllerReference(&dummy, &pod, r.Scheme); err != nil {
			log.Error(err, "Failed to set owner reference to the pod")
			return ctrl.Result{}, err
		}

		// update the pod
		if err := r.Update(ctx, &pod); err != nil {
			log.Error(err, "Failed to update the pod")
			return ctrl.Result{}, err
		}
	}

	if err := r.updateDummyStatus(ctx, &dummy, &pod); err != nil {
		log.Error(err, "Failed to update the dummy status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// createNginxPod creates the nginx pod with the same name and namespace as the dummy object
func (r *DummyReconciler) createNginxPod(dummy *interviewv1alpha1.Dummy) (*corev1.Pod, error) {
	pod := corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      dummy.Name,
			Namespace: dummy.Namespace,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  "nginx-container",
					Image: "nginx:latest",
				},
			},
		},
	}

	// set the owner reference to the dummy resource
	if err := controllerutil.SetControllerReference(dummy, &pod, r.Scheme); err != nil {
		return nil, err
	}

	return &pod, nil
}

// updateDummyStatus update the status of dummy resource
func (r *DummyReconciler) updateDummyStatus(ctx context.Context, dummy *interviewv1alpha1.Dummy, pod *corev1.Pod) error {
	// update the dummy status with spec.Message
	if dummy.Status.SpecEcho != dummy.Spec.Message {
		dummy.Status.SpecEcho = dummy.Spec.Message
	}

	// update the dummy status with the pod status
	if dummy.Status.PodStatus != string(pod.Status.Phase) {
		dummy.Status.PodStatus = string(pod.Status.Phase)
	}

	err := r.Status().Update(ctx, dummy)
	if err != nil {
		return err
	}

	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *DummyReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&interviewv1alpha1.Dummy{}).
		Owns(&corev1.Pod{}). // feed the controlled pod changes to the controller
		Complete(r)
}
