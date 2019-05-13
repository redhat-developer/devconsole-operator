package topology

import (
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
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

	// coreClient is kubernetes go client which gets intialized using mgr.Config().
	coreClient *kubernetes.Clientset
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

	// Initialize kubernetes client
	cl, err := kubernetes.NewForConfig(mgr.GetConfig())
	if err != nil {
		log.Error(err, "Failed to create rest client")
		return &ReconcileService{client: mgr.GetClient(), scheme: mgr.GetScheme(), coreClient: cl}
	}

	// Check if Deployment already exists
	_, err = cl.AppsV1().Deployments(ServicesNamespace).Get(ServiceName, metav1.GetOptions{})
	if err != nil {
		// If Deployment didn't exist, then we need to create one.
		_, err = cl.AppsV1().Deployments(ServicesNamespace).Create(newDeploymentConfigForAppService(nil, ServiceName, ServicesNamespace))
		if err != nil {
			log.Error(err, "Failed to create deployment")
		}
	}

	// Moving ahead to create service assuming deployment is created succesfully.
	_, err = cl.CoreV1().Services(ServicesNamespace).Get(ServiceName, metav1.GetOptions{})
	if err != nil {
		svc, _ := newService(8080)
		_, err = cl.CoreV1().Services(ServicesNamespace).Create(svc)
		if err != nil {
			log.Error(err, "Failed to create service")
		}
	}

	return &ReconcileService{client: mgr.GetClient(), scheme: mgr.GetScheme(), coreClient: cl}
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

	// Watch for Deployment Update and Delete event
	err = c.Watch(&source.Kind{Type: &appsv1.Deployment{}}, &handler.EnqueueRequestForObject{}, pred)
	if err != nil {
		return err
	}

	// Watch for Service Update and Delete event
	err = c.Watch(&source.Kind{Type: &corev1.Service{}}, &handler.EnqueueRequestForObject{}, pred)
	if err != nil {
		return err
	}
	return nil
}

// Reconcile handles events related to changes to the App Topology Service deployment.
// This includes events from service/dc named "ServiceName" in the namespace "ServiceNameSpace"
func (r *ReconcileService) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	// Check if deployment exist or not, create one if absent
	dExist, err := r.coreClient.AppsV1().Deployments(ServicesNamespace).Get(ServiceName, metav1.GetOptions{})
	if dExist.Name == ServiceName {
		log.Info("Deployment already exist with the name : %s", dExist.Name)
	}
	if err != nil {
		_, err = r.coreClient.AppsV1().Deployments(ServicesNamespace).Create(newDeploymentConfigForAppService(nil, ServiceName, ServicesNamespace))
		if err != nil {
			log.Error(err, "Failed to redeploy deployment")
			return reconcile.Result{}, err
		}
	}

	// Check if service exist or not, create one if absent
	svcExist, err := r.coreClient.CoreV1().Services(ServicesNamespace).Get(ServiceName, metav1.GetOptions{})
	if svcExist.Name == ServiceName {
		return reconcile.Result{}, err
	}
	if err != nil {
		newSvc, err := newService(8080)
		if err != nil {
			return reconcile.Result{}, err
		}
		_, err = r.coreClient.CoreV1().Services(ServicesNamespace).Create(newSvc)
		if err != nil {
			fmt.Println("Failed to redeploy dc")
		}
	}

	return reconcile.Result{}, nil
}

func newDeploymentConfigForAppService(containerPorts []corev1.ContainerPort, serviceName string, serviceNameSpace string) *appsv1.Deployment {
	labels := getLabelsForServiceDeployments(ServiceName)
	//annotations := resource.GetAnnotationsForCR(cp)
	if containerPorts == nil {
		containerPorts = []corev1.ContainerPort{{
			ContainerPort: 8080,
			Protocol:      corev1.ProtocolTCP,
		}}
	}
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ServiceName,
			Namespace: ServicesNamespace,
			Labels:    labels,
		},
		Spec: appsv1.DeploymentSpec{
			Strategy: appsv1.DeploymentStrategy{
				Type: "Recreate",
			},
			Replicas: int32Ptr(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name:      ServiceName,
					Namespace: ServicesNamespace,
					Labels:    labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  ServiceName,
							Image: "quay.io/redhat-developer/app-service:latest", // TODO(Akash): parameterize this
							Ports: containerPorts,
						},
					},
				},
			},
		},
	}
}

func newService(port int32) (*corev1.Service, error) {
	labels := getLabelsForServiceDeployments(ServiceName)
	if port > 65536 || port < 1024 {
		return nil, fmt.Errorf("port %d is out of range [1024-65535]", port)
	}
	var svcPorts []corev1.ServicePort
	svcPort := corev1.ServicePort{
		Name:       ServiceName + "-tcp",
		Port:       port,
		Protocol:   corev1.ProtocolTCP,
		TargetPort: intstr.FromInt(int(port)),
	}
	svcPorts = append(svcPorts, svcPort)
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ServiceName,
			Namespace: ServicesNamespace,
			Labels:    labels,
		},
		Spec: corev1.ServiceSpec{
			Ports: svcPorts,
			Selector: map[string]string{
				"deploymentconfig": ServiceName,
			},
		},
	}
	return svc, nil
}

func getLabelsForServiceDeployments(serviceName string) map[string]string {
	labels := make(map[string]string)
	labels["app.kubernetes.io/name"] = serviceName
	labels["app"] = serviceName

	return labels
}

func int32Ptr(i int32) *int32 { return &i }
