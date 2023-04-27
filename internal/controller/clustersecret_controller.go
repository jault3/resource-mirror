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

package controller

import (
	"context"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	resourcemirrorv1alpha1 "github.com/jault3/resource-mirror/api/v1alpha1"
)

// ClusterSecretReconciler reconciles a ClusterSecret object
type ClusterSecretReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=resourcemirror.joshault.dev,resources=clustersecrets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=resourcemirror.joshault.dev,resources=clustersecrets/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=resourcemirror.joshault.dev,resources=clustersecrets/finalizers,verbs=update
//+kubebuilder:rbac:groups=core,resources=secrets,verbs=list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the ClusterSecret object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.4/pkg/reconcile
func (r *ClusterSecretReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	var cs resourcemirrorv1alpha1.ClusterSecret
	if err := r.Get(ctx, req.NamespacedName, &cs); err != nil {
		log.Error(err, "unable to fetch ClusterSecret")
		// we'll ignore not-found errors, since they can't be fixed by an immediate
		// requeue (we'll need to wait for a new notification), and we can get them
		// on deleted requests.
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	var namespaces corev1.NamespaceList
	if err := r.List(ctx, &namespaces); err != nil {
		log.Error(err, "unable to fetch Namespaces to mirror to")
		return ctrl.Result{}, err
	}

	mirroredTo := []string{}
	for _, ns := range namespaces.Items {
		copy := cs.DeepCopy()

		secret := corev1.Secret{}
		secret.Name = copy.Name
		secret.Namespace = ns.Name
		secret.Labels = copy.Labels
		secret.Annotations = copy.Annotations
		secret.Type = cs.Spec.Type
		secret.Data = copy.Spec.Data

		err := ctrl.SetControllerReference(&cs, &secret, r.Scheme)
		if err != nil {
			log.Error(err, "failed to set owner references on child secret for namespace", "namespace", ns.Name)
			return ctrl.Result{}, err
		}

		result, err := ctrl.CreateOrUpdate(ctx, r.Client, &secret, func() error {
			secret.Labels = copy.Labels
			secret.Annotations = copy.Annotations
			secret.Type = cs.Spec.Type
			secret.Data = copy.Spec.Data
			return nil
		})
		if err != nil {
			log.Error(err, "failed to create or update child secret in namespace", "namespace", ns.Name)
			return ctrl.Result{}, err
		}
		log.Info(fmt.Sprintf("%s/%s %s", ns.Name, secret.Name, result))

		mirroredTo = append(mirroredTo, ns.Name)
	}

	cs.Status.Mirrored = true
	cs.Status.MirroredTo = mirroredTo
	cs.Status.LastReconciled = time.Now().Format(time.RFC3339)

	if err := r.Status().Update(ctx, &cs); err != nil {
		log.Error(err, "unable to update ClusterSecret status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ClusterSecretReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&resourcemirrorv1alpha1.ClusterSecret{}).
		Owns(&corev1.Secret{}).
		WithEventFilter(predicate.GenerationChangedPredicate{}).
		Complete(r)
}
