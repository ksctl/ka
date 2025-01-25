/*
Copyright 2025.

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
	"slices"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	appv1 "github.com/ksctl/ka/api/v1"
)

// StackReconciler reconciles a Stack object
type StackReconciler struct {
	client.Client
	RestConfig *rest.Config
	Scheme     *runtime.Scheme
	state      *StackState
}

func (r *StackReconciler) InitializeStorage(ctx context.Context) error {
	if r.state == nil {
		return r.Load(ctx)
	}
	return nil
}

const managerFinalizer string = "finalizer.stack.app.ksctl.com"

// +kubebuilder:rbac:groups=app.ksctl.com,resources=stacks,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=app.ksctl.com,resources=stacks/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=app.ksctl.com,resources=stacks/finalizers,verbs=update
// +kubebuilder:rbac:groups=*,resources=*,verbs=get;list;watch;create;update;patch;delete

func (r *StackReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx)
	l.Info("Reconciling Stack", "stack", req.NamespacedName)

	if err := r.InitializeStorage(ctx); err != nil {
		l.Error(err, "Failed to initialize storage")
		return ctrl.Result{RequeueAfter: time.Second * 5}, err
	}

	instance := &appv1.Stack{}

	if err := r.Get(ctx, req.NamespacedName, instance); err != nil {
		if client.IgnoreNotFound(err) == nil {
			return ctrl.Result{}, nil // Object deleted, no requeue
		}
		l.Error(err, "Failed to get ClusterAddon")
		return ctrl.Result{RequeueAfter: time.Second * 10}, err
	}

	if instance.Status.StatusCode == "" {
		instance.Status.StatusCode = appv1.WorkingOn
		if err := r.Status().Update(ctx, instance); err != nil {
			l.Error(err, "Failed to update initial status")
			return ctrl.Result{RequeueAfter: time.Second * 5}, err
		}
	}

	if !instance.DeletionTimestamp.IsZero() {
		return r.processDeletion(ctx, instance)
	}

	if !slices.Contains(instance.Finalizers, managerFinalizer) {
		return r.addFinalizer(ctx, instance)
	}

	return r.processInstall(ctx, instance)
}

func (r *StackReconciler) processInstall(ctx context.Context, instance *appv1.Stack) (ctrl.Result, error) {
	l := log.FromContext(ctx)

	if err := r.Add(ctx, instance); err != nil {
		l.Error(err, "Failed to install app")

		instance.Status.StatusCode = appv1.Failure
		instance.Status.ReasonOfFailure = err.Error() + "\nFailed to install app"

		if err := r.Status().Update(ctx, instance); err != nil {
			l.Error(err, "Failed to update status")
			return ctrl.Result{RequeueAfter: time.Second * 5}, err
		}

		return ctrl.Result{RequeueAfter: time.Second * 5}, err
	}

	instance.Status.StatusCode = appv1.Success
	instance.Status.ReasonOfFailure = ""
	if err := r.Status().Update(ctx, instance); err != nil {
		l.Error(err, "Failed to update status")
		return ctrl.Result{RequeueAfter: time.Second * 5}, err
	}
	return ctrl.Result{}, nil
}

func (r *StackReconciler) processDeletion(ctx context.Context, instance *appv1.Stack) (ctrl.Result, error) {
	l := log.FromContext(ctx)

	if !slices.Contains(instance.Finalizers, managerFinalizer) {
		l.Info("Finalizer already removed")
		return ctrl.Result{}, nil
	}

	if err := r.Remove(ctx, instance); err != nil {
		l.Error(err, "Failed to uninstall app")
		return ctrl.Result{RequeueAfter: time.Second * 5}, err
	}

	if _, err := r.removeFinalizer(ctx, instance); err != nil {
		l.Error(err, "Failed to remove finalizer")
		return ctrl.Result{RequeueAfter: time.Second * 5}, err
	}

	return ctrl.Result{}, nil
}

func (r *StackReconciler) addFinalizer(ctx context.Context, instance *appv1.Stack) (ctrl.Result, error) {
	instance.Finalizers = append(instance.Finalizers, managerFinalizer)
	if err := r.Update(ctx, instance); err != nil {
		return ctrl.Result{RequeueAfter: time.Second * 5}, err
	}
	return ctrl.Result{Requeue: true}, nil
}

func (r *StackReconciler) removeFinalizer(ctx context.Context, instance *appv1.Stack) (ctrl.Result, error) {
	if !slices.Contains(instance.Finalizers, managerFinalizer) {
		return ctrl.Result{}, nil
	}

	var v []string

	for _, f := range instance.Finalizers {
		if f != managerFinalizer {
			v = append(v, f)
		}
	}

	instance.Finalizers = v
	if err := r.Update(ctx, instance); err != nil {
		return ctrl.Result{RequeueAfter: time.Second * 5}, err
	}
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *StackReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appv1.Stack{}).
		Named("stack").
		Complete(r)
}
