package component

import (
	"context"

	devopsconsolev1alpha1 "github.com/redhat-developer/devopsconsole-operator/pkg/apis/devopsconsole/v1alpha1"

	v1 "github.com/openshift/api/apps/v1"
	buildv1 "github.com/openshift/api/build/v1"
	imagev1 "github.com/openshift/api/image/v1"

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

	build "github.com/openshift/api/build/v1"

	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_component")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new Component Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	sc := mgr.GetScheme()
	build.AddToScheme(sc)
	imagev1.AddToScheme(sc)
	v1.AddToScheme(sc)

	return &ReconcileComponent{client: mgr.GetClient(), scheme: sc}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("component-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource Component
	err = c.Watch(&source.Kind{Type: &devopsconsolev1alpha1.Component{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner Component
	err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &devopsconsolev1alpha1.Component{},
	})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &imagev1.ImageStream{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &devopsconsolev1alpha1.Component{},
	})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &buildv1.BuildConfig{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &devopsconsolev1alpha1.Component{},
	})
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
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileComponent) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling Component - because of primary resource or a secondary resource")

	// Fetch the Component instance
	instance := &devopsconsolev1alpha1.Component{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			reqLogger.Info("Got triggered when owned objects are being deleted. Looks like primary resource is gone!")
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	// Define a new Pod object
	pod := newPodForCR(instance)

	// Set Component instance as the owner and controller
	if err := controllerutil.SetControllerReference(instance, pod, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	// Check if this Pod already exists
	found := &corev1.Pod{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: pod.Name, Namespace: pod.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a new Pod", "Pod.Namespace", pod.Namespace, "Pod.Name", pod.Name)
		err = r.client.Create(context.TODO(), pod)
		if err != nil {
			return reconcile.Result{}, err
		}

		// Pod created successfully - don't requeue
		return reconcile.Result{}, nil
	} else if err != nil {
		return reconcile.Result{}, err
	}

	newBuildConfig := newBuildConfigForCR(instance)
	foundBuildConfig := &buildv1.BuildConfig{}

	// Set Component instance as the owner and controller
	if err = controllerutil.SetControllerReference(instance, newBuildConfig, r.scheme); err != nil {
		reqLogger.Error(err, "Couldn't set controller owner")
		return reconcile.Result{}, err
	}

	err = r.client.Get(context.TODO(), types.NamespacedName{Name: newBuildConfig.Name, Namespace: newBuildConfig.Namespace}, foundBuildConfig)
	if err != nil && errors.IsNotFound(err) {

		reqLogger.Info("Creating a new BuildConfig", "Buildconfig.Namespace", newBuildConfig.Namespace, "BuildConfig.Name", newBuildConfig.Name)

		err = r.client.Create(context.TODO(), newBuildConfig)
		if err != nil {
			return reconcile.Result{}, err
		}

		// BuildConfig created successfully - don't requeue
		reqLogger.Info("BuildConfig created", "BuildConfig.Namespace", foundBuildConfig.Namespace, "BuildConfig.Name", foundBuildConfig.Name)
		return reconcile.Result{}, nil

	} else if err != nil {
		return reconcile.Result{}, err
	}

	newIS := newImageStream(instance)
	foundImageStream := &imagev1.ImageStream{}

	// Set Component instance as the owner and controller
	if err = controllerutil.SetControllerReference(instance, newIS, r.scheme); err != nil {
		reqLogger.Error(err, "Couldn't set controller owner")
		return reconcile.Result{}, err
	}

	//reqLogger.Info("Checking if image stream exists", "ImageStream.Namespace", found.Namespace, "Pod.Name", found.Name)
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: newIS.Name, Namespace: newIS.Namespace}, foundImageStream)
	if err != nil && errors.IsNotFound(err) {

		reqLogger.Info("Creating a new ImageStream", "ImageStream.Namespace", newIS.Namespace, "BuildConfig.Name", newIS.Name)

		err = r.client.Create(context.TODO(), newIS)
		if err != nil {
			return reconcile.Result{}, err
		}

		reqLogger.Info("ImageStream created", "ImageStream.Namespace", foundImageStream.Namespace, "ImageStream.Name", foundImageStream.Name)

		return reconcile.Result{}, nil

	} else if err != nil {
		return reconcile.Result{}, err
	}

	newDeploymentConfig := newDeploymentConfigForCR(instance)
	foundDeploymentConfig := &v1.DeploymentConfig{}

	// Set Component instance as the owner and controller
	if err = controllerutil.SetControllerReference(instance, newDeploymentConfig, r.scheme); err != nil {
		reqLogger.Error(err, "Couldn't set controller owner")
		return reconcile.Result{}, err
	}

	err = r.client.Get(context.TODO(), types.NamespacedName{Name: newDeploymentConfig.Name, Namespace: newDeploymentConfig.Namespace}, newDeploymentConfig)
	if err != nil && errors.IsNotFound(err) {

		reqLogger.Info("Creating a new dc", "DeploymentConfig.Namespace", newDeploymentConfig.Namespace, "DeploymentConfig.Name", newDeploymentConfig.Name)

		err = r.client.Create(context.TODO(), newDeploymentConfig)
		if err != nil {
			return reconcile.Result{}, err
		}

		reqLogger.Info("DC created", "DeploymentConfig.Namespace", newDeploymentConfig.Namespace, "DeploymentConfig.Name", foundDeploymentConfig.Name)
		return reconcile.Result{}, nil

	} else if err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

func newBuildConfigForCR(cr *devopsconsolev1alpha1.Component) *buildv1.BuildConfig {
	labels := map[string]string{
		"app": cr.Name,
	}
	c := cr.Spec.DeepCopy()
	return &buildv1.BuildConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name + "-buildconfig",
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Spec: buildv1.BuildConfigSpec{
			CommonSpec: buildv1.CommonSpec{
				Output: buildv1.BuildOutput{
					To: &corev1.ObjectReference{
						Kind: "ImageStreamTag",
						Name: cr.Name + ":latest",
					},
				},
				Source: buildv1.BuildSource{
					Git: &buildv1.GitBuildSource{
						URI: c.Codebase,
						Ref: "master",
					},
					Type: "Git",
				},
				Strategy: buildv1.BuildStrategy{
					SourceStrategy: &buildv1.SourceBuildStrategy{
						From: corev1.ObjectReference{
							Kind:      "ImageStreamTag",
							Name:      "nodejs:latest",
							Namespace: "openshift",
						},
					},
				},
			},
		},
	}
}

func newDeploymentConfigForCR(cr *devopsconsolev1alpha1.Component) *v1.DeploymentConfig {
	labels := map[string]string{
		"app": cr.Name,
	}
	return &v1.DeploymentConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name,
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Spec: v1.DeploymentConfigSpec{
			Strategy: v1.DeploymentStrategy{
				Type: v1.DeploymentStrategyTypeRolling,
			},
			Replicas: 1,
			Template: &corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name:      cr.Name,
					Namespace: cr.Namespace,
					Labels:    labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						corev1.Container{
							Name:  cr.Name,
							Image: cr.Name + ":latest",
							Ports: []corev1.ContainerPort{
								corev1.ContainerPort{
									ContainerPort: 8080,
									Protocol:      corev1.ProtocolTCP,
								},
							},
						},
					},
				},
			},
			Triggers: []v1.DeploymentTriggerPolicy{
				v1.DeploymentTriggerPolicy{
					Type: v1.DeploymentTriggerOnConfigChange,
				},
				v1.DeploymentTriggerPolicy{
					Type: v1.DeploymentTriggerOnImageChange,
					ImageChangeParams: &v1.DeploymentTriggerImageChangeParams{
						ContainerNames: []string{
							cr.Name,
						},
						From: corev1.ObjectReference{
							Kind: "ImageStreamTag",
							Name: cr.Name + ":latest",
						},
					},
				},
			},
		},
	}
}

func newImageStream(cr *devopsconsolev1alpha1.Component) *imagev1.ImageStream {
	labels := map[string]string{
		"app": cr.Name,
	}
	return &imagev1.ImageStream{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name,
			Namespace: cr.Namespace,
			Labels:    labels,
		},
	}

}

// newPodForCR returns a busybox pod with the same name/namespace as the cr
func newPodForCR(cr *devopsconsolev1alpha1.Component) *corev1.Pod {
	labels := map[string]string{
		"app": cr.Name,
	}
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name,
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:    "busybox",
					Image:   "busybox",
					Command: []string{"sleep", "3600"},
				},
			},
		},
	}
}
