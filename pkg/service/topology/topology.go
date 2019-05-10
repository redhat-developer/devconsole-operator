package topology

import (
	"fmt"

	v1 "github.com/openshift/api/apps/v1"
	deploymentconfig "github.com/openshift/client-go/apps/clientset/versioned/typed/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log

// ReconcileService reconciles a Component object
type ReconcileService struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Add creates a new Component Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

const (
	// ServicesNamespace is the name of the namespace where this operator would install the Rest Service
	ServicesNamespace = "openshift-operators" // move this out to env var ?

	// ServiceName is the name that would be assigned to all objects associated with the Rest Service
	ServiceName = "devconsole-app" // move this out to env var ?
)

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	config := mgr.GetConfig()

	dcClient, _ := deploymentconfig.NewForConfig(config)

	// Check if DC already exists
	existingDC, _ := dcClient.DeploymentConfigs(ServicesNamespace).Get(ServiceName, metav1.GetOptions{})

	if existingDC.Name == ServiceName {
		return &ReconcileService{client: mgr.GetClient(), scheme: mgr.GetScheme()}
	}

	// If DC didn't exist, then we can skip
	_, err := dcClient.DeploymentConfigs(ServicesNamespace).Create(newDeploymentConfigForAppService(nil, ServiceName, ServicesNamespace))
	if err != nil {
		fmt.Println(err) // Log.Error(..)
	}

	// TODO: Create a service, if absent

	// TODO: Create a route, if absent

	return &ReconcileService{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {

	// Create a new controller
	c, err := controller.New("topology-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Do not send events which are related to the specific DC.
	pred := predicate.Funcs{
		// TODO: When the deployment is being created, DC gets updated
		// and that would trigget this. How do we filter out such events?
		UpdateFunc: func(e event.UpdateEvent) bool {
			return e.MetaOld.GetName() == ServiceName && e.MetaOld.GetNamespace() == ServicesNamespace
		},

		// TODO: In all probability, any delele event is interesting to us.
		DeleteFunc: func(e event.DeleteEvent) bool {
			return e.Meta.GetName() == ServiceName && e.Meta.GetNamespace() == ServicesNamespace
		},

		// TODO: When a new one is created because an operator is being deployed..
		CreateFunc: func(e event.CreateEvent) bool {
			return false
		},
	}

	err = c.Watch(&source.Kind{Type: &v1.DeploymentConfig{}}, &handler.EnqueueRequestForObject{}, pred)
	if err != nil {
		return err
	}
	return nil
}

// Reconcile handles events related to changes to the App Topology Service deployment.
// This includes events from service/route/dc named "ServiceName" in the namespace "ServiceNameSpace"
func (r *ReconcileService) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	// TODO: Watch changes to the DeploymentConfig ( .. and the service, and the route )
	// something happened to the specific DC, please react..
	return reconcile.Result{}, nil
}

func newDeploymentConfigForAppService(containerPorts []corev1.ContainerPort, serviceName string, serviceNameSpace string) *v1.DeploymentConfig {
	labels := getLabelsForServiceDeployments(ServiceName)
	//annotations := resource.GetAnnotationsForCR(cp)
	if containerPorts == nil {
		containerPorts = []corev1.ContainerPort{{
			ContainerPort: 8080,
			Protocol:      corev1.ProtocolTCP,
		}}
	}
	return &v1.DeploymentConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ServiceName,
			Namespace: ServicesNamespace,
			Labels:    labels,
			//	Annotations: annotations,
		},
		Spec: v1.DeploymentConfigSpec{
			Strategy: v1.DeploymentStrategy{
				Type: v1.DeploymentStrategyTypeRecreate,
			},
			Replicas: 1,
			Selector: labels,
			Template: &corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name:      ServiceName,
					Namespace: ServicesNamespace,
					Labels:    labels,
					//	Annotations: annotations,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  ServiceName,
							Image: "quay.io/redhat-developer/app-service:latest", // parameterize this
							Ports: containerPorts,
						},
					},
				},
			},
			Triggers: []v1.DeploymentTriggerPolicy{
				{
					Type: v1.DeploymentTriggerOnConfigChange,
				},
			},
		},
	}
}

func getLabelsForServiceDeployments(serviceName string) map[string]string {
	labels := make(map[string]string)
	labels["app.kubernetes.io/name"] = serviceName
	labels["app"] = serviceName

	return labels
}
