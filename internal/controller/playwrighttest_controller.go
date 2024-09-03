/*
Copyright 2024.

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

package controller

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	testingv1 "github.com/JimmyChen233/playwright-operator/api/v1"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
)

// PlaywrightTestReconciler reconciles a PlaywrightTest object
type PlaywrightTestReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=testing.my.domain,resources=playwrighttests,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=testing.my.domain,resources=playwrighttests/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=testing.my.domain,resources=playwrighttests/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the PlaywrightTest object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.17.3/pkg/reconcile
func (r *PlaywrightTestReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("playwrighttest", req.NamespacedName)
	// fetch playwright instance
	instance := &testingv1.PlaywrightTest{}
	if err := r.Get(ctx, req.NamespacedName, instance); err != nil {
		log.Error(err, "unable to fetch PlaywrightTest")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	// crd resoruce is marked as delete
	if instance.DeletionTimestamp != nil {
		return ctrl.Result{}, nil
	}
	// Fetch the ConfigMap containing the test configuration
	configMap := &corev1.ConfigMap{}
	if err := r.Get(ctx, types.NamespacedName{Name: instance.Spec.ConfigMap, Namespace: req.Namespace}, configMap); err != nil {
		log.Error(err, "unable to fetch ConfigMap")
		return ctrl.Result{}, err
	}
	// Create a Kubernetes Job to run the Playwright test
	job := createPlaywrightJob(playwrightTest, configMap)
	if err := r.Client.Create(ctx, job); err != nil {
		log.Error(err, "unable to create Job for PlaywrightTest")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *PlaywrightTestReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&testingv1.PlaywrightTest{}).
		Complete(r)
}
