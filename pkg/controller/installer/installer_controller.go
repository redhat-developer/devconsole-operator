package installer

import (
	"context"
	"fmt"
	"os"

	"github.com/go-logr/logr"
	routev1 "github.com/openshift/api/route/v1"
	devopsconsolev1alpha1 "github.com/redhat-developer/devopsconsole-operator/pkg/apis/devopsconsole/v1alpha1"
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

var log = logf.Log.WithName("controller_installer")
var appServiceEnvVar = "APP_SERVICE_IMAGE_NAME"

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new Installer Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileInstaller{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("installer-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource Installer
	err = c.Watch(&source.Kind{Type: &devopsconsolev1alpha1.Installer{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// Watch for changes to secondary resource Deployment and requeue the owner Installer
	err = c.Watch(&source.Kind{Type: &appv1.Deployment{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &devopsconsolev1alpha1.Installer{},
	})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &corev1.Service{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &devopsconsolev1alpha1.Installer{},
	})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &routev1.Route{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &devopsconsolev1alpha1.Installer{},
	})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileInstaller{}

// ReconcileInstaller reconciles a Installer object
type ReconcileInstaller struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client    client.Client
	scheme    *runtime.Scheme
	reqLogger logr.Logger
}

// Reconcile reads that state of the cluster for a Installer object and makes changes based on the state read
// and what is in the Installer.Spec
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileInstaller) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	r.reqLogger = log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	r.reqLogger.Info("Reconciling Installer")

	// Fetch the Installer instance
	instance := &devopsconsolev1alpha1.Installer{}
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

	if _, err = r.createDeployment(instance); err != nil {
		return reconcile.Result{}, err
	}

	service, err := r.createService(instance)
	if err != nil {
		return reconcile.Result{}, err
	}

	route, err := r.createRoute(instance, service.Name)
	if err != nil {
		return reconcile.Result{}, err
	}

	r.reqLogger.Info("Created Deployment, Service and Route")

	if err := r.updateStatus(instance, route); err != nil {
		return reconcile.Result{}, err
	}

	r.reqLogger.Info("Reconciliation completed.")
	return reconcile.Result{}, nil
}

// createDeployment creates a new deployment if it doesn't already exists.
// Returns the old deployment if found, or the new deployment if created.
func (r *ReconcileInstaller) createDeployment(cr *devopsconsolev1alpha1.Installer) (*appv1.Deployment, error) {
	// Define a new deployment object
	deployment, err := newDeploymentForCR(cr, 1)
	if err != nil {
		return nil, err
	}
	// Set Installer cr as the owner and controller
	if err := controllerutil.SetControllerReference(cr, deployment, r.scheme); err != nil {
		return nil, err
	}

	foundDeployment := &appv1.Deployment{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: deployment.Name, Namespace: deployment.Namespace}, foundDeployment)

	// A deployment already exists. Don't do anything
	if err == nil {
		r.reqLogger.Info("Skip Creating Deployment: Deployment Already exist", "Deployment.Namespace", foundDeployment.Namespace, "Deployment.Name", foundDeployment.Name)
		return foundDeployment, nil
	}

	// A deployment doesn't exist. Create a new one
	if errors.IsNotFound(err) {
		r.reqLogger.Info("Creating a new Deployment", "Deployment.Namespace", deployment.Namespace, "Deployment.Name", deployment.Name)
		err = r.client.Create(context.TODO(), deployment)
		if err != nil && !errors.IsAlreadyExists(err) {
			return nil, err
		}
		r.reqLogger.Info("Successfully created new deployment", "Deployment.Namespace", deployment.Namespace, "Deployment.Name", deployment.Name)
		return deployment, nil
	}

	// Something else went wrong while trying to load the existing deployment
	return nil, err
}

// createService creates a new Service if it doesn't already exists.
// Returns the old Service if found, or the new Service if created.
func (r *ReconcileInstaller) createService(cr *devopsconsolev1alpha1.Installer) (*corev1.Service, error) {
	newService := newServiceForCR(cr)
	// Set Installer cr as the owner and controller
	if err := controllerutil.SetControllerReference(cr, newService, r.scheme); err != nil {
		return nil, err
	}

	foundService := &corev1.Service{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: newService.Name, Namespace: newService.Namespace}, foundService)

	// A service already exists. Don't do anything
	if err == nil {
		r.reqLogger.Info("Skip Creating Service: Service Already exist", "Service.Namespace", foundService.Namespace, "Service.Name", foundService.Name)
		return foundService, nil
	}

	// The service doesn't exist. Create a new one
	if errors.IsNotFound(err) {
		r.reqLogger.Info("Creating a new Service", "Service.Namespace", newService.Namespace, "Service.Name", newService.Name)
		err = r.client.Create(context.TODO(), newService)
		if err != nil && !errors.IsAlreadyExists(err) {
			return nil, err
		}
		r.reqLogger.Info("Successfully created new Service", "Service.Namespace", newService.Namespace, "Service.Name", newService.Name)
		return newService, nil
	}

	// Something else went wrong. Return the error and try again
	return nil, err

}

// createRoute creates a new Route if it doesn't already exists.
// Returns the old Route if found, or the new Route if created.
func (r *ReconcileInstaller) createRoute(cr *devopsconsolev1alpha1.Installer, serviceName string) (*routev1.Route, error) {
	newRoute := newRouteForCR(cr, serviceName)
	// Set Installer cr as the owner and controller
	if err := controllerutil.SetControllerReference(cr, newRoute, r.scheme); err != nil {
		return nil, err
	}

	foundRoute := &routev1.Route{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: newRoute.Name, Namespace: newRoute.Namespace}, foundRoute)
	// Route already exists. Don't do anything
	if err == nil {
		r.reqLogger.Info("Skip Creating Route: Route Already exist", "Route.Namespace", foundRoute.Namespace, "Route.Name", foundRoute.Name)
		return foundRoute, nil
	}

	// Route doesn't exist. Try to create a new one
	if errors.IsNotFound(err) {
		r.reqLogger.Info("Creating a new Route", "Route.Namespace", newRoute.Namespace, "Route.Name", newRoute.Name)
		err = r.client.Create(context.TODO(), newRoute)
		if err != nil && errors.IsAlreadyExists(err) {
			// A route was created by a different event while were reconcilling.
			// Skip creation.
			r.reqLogger.Info("Route already exists", "Route.Namespace", newRoute.Namespace, "Route.Name", newRoute.Name)
			return newRoute, nil
		}
		if err != nil {
			r.reqLogger.Error(err, "Failed to create new route.")
			return nil, err
		}
		r.reqLogger.Info("Successfully created new Route", "Route.Namespace", newRoute.Namespace, "Route.Name", newRoute.Name)
		return newRoute, nil
	}
	return nil, err
}

// updateStatus adds update the cr.Status.AppServiceURL to reflect the app
// service url
func (r *ReconcileInstaller) updateStatus(cr *devopsconsolev1alpha1.Installer, route *routev1.Route) error {
	if route.Spec.Host == cr.Status.AppServiceURL {
		r.reqLogger.Info("Skip CR status Update: CR status already up-to-date", "Route.Name", route.Name, "Route.Status.AppServiceURL", cr.Status.AppServiceURL)
		return nil
	}

	// Update the status
	r.reqLogger.Info("Updating CR Status", "current route", cr.Status.AppServiceURL, "new route", route.Spec.Host)
	cr.Status.AppServiceURL = route.Spec.Host
	if err := r.client.Status().Update(context.TODO(), cr); err != nil {
		r.reqLogger.Error(err, "Failed to update Installer status.")
		return err
	}

	r.reqLogger.Info("Successfully updated CR status", "CR.Namespace", cr.Name, "CR.Status.AppServiceURL", cr.Status.AppServiceURL)
	return nil
}

func getLabelsForCR(cr *devopsconsolev1alpha1.Installer) map[string]string {
	labels := map[string]string{
		"app": cr.Name,
	}
	return labels
}

// newDeploymentForCR returns a deployment with the same name/namespace as the cr
func newDeploymentForCR(cr *devopsconsolev1alpha1.Installer, replicas int32) (*appv1.Deployment, error) {
	imageName, err := GetEnvValue(appServiceEnvVar)
	if err != nil {
		return nil, err
	}
	labels := getLabelsForCR(cr)
	return &appv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "appservice-deploy",
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
	}, nil
}

func newServiceForCR(cr *devopsconsolev1alpha1.Installer) *corev1.Service {
	labels := getLabelsForCR(cr)
	svc := &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "appservice-service",
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

func newRouteForCR(cr *devopsconsolev1alpha1.Installer, serviceName string) *routev1.Route {
	labels := getLabelsForCR(cr)
	route := &routev1.Route{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Route",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "appservice-route",
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Spec: routev1.RouteSpec{
			To: routev1.RouteTargetReference{
				Kind: "Service",
				Name: serviceName,
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
