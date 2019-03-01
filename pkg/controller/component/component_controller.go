package component

import (
	"context"
	"fmt"
	"strconv"

	componentsv1alpha1 "github.com/redhat-developer/devopsconsole-operator/pkg/apis/components/v1alpha1"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_component")

// Add creates a new Component Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileComponent{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("component-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource Component
	err = c.Watch(&source.Kind{Type: &componentsv1alpha1.Component{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileComponent{}

// ReconcileComponent reconciles a Component object
type ReconcileComponent struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a Component object and makes changes based on the state read
// and what is in the Component.Spec
func (r *ReconcileComponent) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling Component")

	// Fetch the Component instance
	instance := &componentsv1alpha1.Component{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	log.Info("============================================================")
	log.Info(fmt.Sprintf("***** Reconciling Component %s, namespace %s", request.Name, request.Namespace))
	log.Info(fmt.Sprintf("** Creation time : %s", instance.ObjectMeta.CreationTimestamp))
	log.Info(fmt.Sprintf("** Resource version : %s", instance.ObjectMeta.ResourceVersion))
	log.Info(fmt.Sprintf("** Generation version : %s", strconv.FormatInt(instance.ObjectMeta.Generation, 10)))
	log.Info(fmt.Sprintf("** Deletion time : %s", instance.ObjectMeta.DeletionTimestamp))
	log.Info("============================================================")

	return reconcile.Result{}, nil
}

