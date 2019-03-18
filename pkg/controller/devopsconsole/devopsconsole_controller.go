package devopsconsole

import (
	"context"
	"fmt"
	"os"

	devopsconsolev1alpha1 "github.com/redhat-developer/devopsconsole-operator/pkg/apis/devopsconsole/v1alpha1"

	routev1 "github.com/openshift/api/route/v1"
	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_devopsconsole")
var appServiceEnvVar = "APP_SERVICE_IMAGE_NAME"
var appServiceCommandEnvVar = "APP_SERVICE_START_COMMAND"

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new DevopsConsole Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileDevopsConsole{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("devopsconsole-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource DevopsConsole
	err = c.Watch(&source.Kind{Type: &devopsconsolev1alpha1.DevopsConsole{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// Watch for changes to secondary resource Deployment and requeue the owner DevopsConsole
	err = c.Watch(&source.Kind{Type: &appv1.Deployment{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &devopsconsolev1alpha1.DevopsConsole{},
	})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &corev1.Service{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &devopsconsolev1alpha1.DevopsConsole{},
	})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &routev1.Route{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &devopsconsolev1alpha1.DevopsConsole{},
	})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileDevopsConsole{}

// ReconcileDevopsConsole reconciles a DevopsConsole object
type ReconcileDevopsConsole struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a DevopsConsole object and makes changes based on the state read
// and what is in the DevopsConsole.Spec
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileDevopsConsole) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling DevopsConsole")

	// Fetch the DevopsConsole instance
	instance := &devopsconsolev1alpha1.DevopsConsole{}
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

	var imageName string
	if imageName, err = GetEnvValue(appServiceEnvVar); err != nil {
		return reconcile.Result{}, err
	}

	// Define a new deployment object
	deployment := newDeploymentForCR(instance, imageName, 1)

	// Set DevopsConsole instance as the owner and controller
	if err := controllerutil.SetControllerReference(instance, deployment, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	// Check if this deployment already exists
	foundDeployment := &appv1.Deployment{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: deployment.Name, Namespace: deployment.Namespace}, foundDeployment)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a new Deployment", "Deployment.Namespace", deployment.Namespace, "Deployment.Name", deployment.Name)
		err = r.client.Create(context.TODO(), deployment)
		if err != nil {
			return reconcile.Result{}, err
		}
		// Deployment created successfully - don't requeue
		return reconcile.Result{}, nil
	} else if err != nil {
		return reconcile.Result{}, err
	}

	// Check if this deployment already exists
	newService := newService(instance)
	// Set DevopsConsole instance as the owner and controller
	if err := controllerutil.SetControllerReference(instance, newService, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	foundService := &corev1.Service{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: newService.Name, Namespace: newService.Namespace}, foundService)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a new Service", "Service.Namespace", newService.Namespace, "Service.Name", newService.Name)
		err = r.client.Create(context.TODO(), newService)
		if err != nil {
			return reconcile.Result{}, err
		}
		// Deployment created successfully - don't requeue
		return reconcile.Result{}, nil
	} else if err != nil {
		return reconcile.Result{}, err
	}

	newRoute := newRoute(instance, newService)
	// Set DevopsConsole instance as the owner and controller
	if err := controllerutil.SetControllerReference(instance, newRoute, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	foundRoute := &routev1.Route{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: newRoute.Name, Namespace: newRoute.Namespace}, foundRoute)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a new Route", "Route.Namespace", newRoute.Namespace, "Route.Name", newRoute.Name)
		err = r.client.Create(context.TODO(), newRoute)
		if err != nil {
			return reconcile.Result{}, err
		}
		// Deployment created successfully - don't requeue
		return reconcile.Result{}, nil
	} else if err != nil {
		return reconcile.Result{}, err
	}

	// Deployment already exists - don't requeue
	reqLogger.Info("Done reconcile: Deployment, Service and Route exists")

	return reconcile.Result{}, nil
}

// newDeploymentForCR returns a busybox pod with the same name/namespace as the cr
func newDeploymentForCR(cr *devopsconsolev1alpha1.DevopsConsole, imageName string, replicas int32) *appv1.Deployment {
	labels := map[string]string{
		"app": cr.Name,
	}
	return &appv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name + "-deploy",
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Spec: appv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Image: imageName,
						Name:  "app-service",
					}},
				},
			},
		},
	}
}

func newService(cr *devopsconsolev1alpha1.DevopsConsole) *corev1.Service {
	labels := map[string]string{
		"app": cr.Name,
	}
	svc := &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "devopsconsoleservice",
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Spec: corev1.ServiceSpec{
			Selector: labels,
			Ports: []corev1.ServicePort{
				{
					Name:     "main",
					Protocol: corev1.ProtocolTCP,
					Port:     8080,
				},
			},
		},
	}
	return svc
}

func newRoute(cr *devopsconsolev1alpha1.DevopsConsole, svc *corev1.Service) *routev1.Route {
	labels := map[string]string{
		"app": cr.Name,
	}
	route := &routev1.Route{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Route",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "devopsconsoleroute",
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Spec: routev1.RouteSpec{
			To: routev1.RouteTargetReference{
				Kind: "Service",
				Name: svc.Name,
			},
		},
	}
	return route
}

// GetEnvValue returns the value of env from the environment. Returns error if
// env is not set.
func GetEnvValue(env string) (string, error) {
	value, found := os.LookupEnv(env)
	if !found {
		return "", fmt.Errorf("%s must be set", env)
	}
	return value, nil
}
