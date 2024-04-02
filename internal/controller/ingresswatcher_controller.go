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

	"fmt"

	"github.com/go-logr/logr"
	"github.com/rbsn-joses/event-watcher/influx_metrics"
	networkv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/runtime"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// IngressWatcherReconciler reconciles a IngressWatcher object
type IngressWatcherReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=my.domain,resources=Ingresswatchers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=my.domain,resources=Ingresswatchers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=my.domain,resources=Ingresswatchers/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the IngressWatcher object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.16.3/pkg/reconcile
var Ctx context.Context
var Logger logr.Logger
var Req ctrl.Request

func (r *IngressWatcherReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	Logger = logger
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *IngressWatcherReconciler) SetupWithManager(mgr ctrl.Manager) error {
	ingress := &networkv1.Ingress{}
	err := r.Get(Ctx, Req.NamespacedName, ingress)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(ingress)
	return ctrl.NewControllerManagedBy(mgr).
		For(ingress).
		WithEventFilter(predicate.Funcs{
			CreateFunc: func(e event.CreateEvent) bool {
				Logger.Info(fmt.Sprintf("Created Ingress %v", Req.NamespacedName))

				influx_metrics.CreateMetrics(e.Object.GetName(), e.Object.GetNamespace(), "CREATE", "ingress", 1)
				return true
			},
			UpdateFunc: func(e event.UpdateEvent) bool {
				Logger.Info(fmt.Sprintf("Updated Ingress %v", Req.NamespacedName))

				// Ignore updates to CR status in which case metadata.Generation does not change
				err := influx_metrics.CreateMetrics(e.ObjectOld.GetName(), e.ObjectNew.GetNamespace(), "UPDATE", "ingress", 2)
				if err != nil {
					fmt.Println(err)
				}
				return e.ObjectOld.GetResourceVersion() != e.ObjectNew.GetResourceVersion()
			},
			DeleteFunc: func(e event.DeleteEvent) bool {
				Logger.Info(fmt.Sprintf("Deleted Ingress %v", e.Object.GetName()+"/"+e.Object.GetNamespace()))

				// Evaluates to false if the object has been confirmed deleted.
				err := influx_metrics.CreateMetrics(e.Object.GetName(), e.Object.GetNamespace(), "DELETE", "ingress", 3)
				if err != nil {
					fmt.Println(err)
				}
				return !e.DeleteStateUnknown
			},
		}).
		Complete(r)

}
