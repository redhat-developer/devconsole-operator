package component

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"strconv"

	gitsourcev1alpha1 "github.com/redhat-developer/devopsconsole-operator/pkg/apis/devopsconsole-operator/v1alpha1"
	componentsv1alpha1 "github.com/redhat-developer/devopsconsole-operator/pkg/apis/devopsconsole/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"

	v1 "github.com/openshift/api/apps/v1"
	buildv1 "github.com/openshift/api/build/v1"
	imagev1 "github.com/openshift/api/image/v1"
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

var (
	_                  reconcile.Reconciler = &ReconcileComponent{}
	buildTypeImages                         = map[string]string{"nodejs": "nodeshift/centos7-s2i-nodejs:10.x"}
	openshiftNamespace                      = "openshift"
)

// ReconcileComponent reconciles a Component object
type ReconcileComponent struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a Component object and makes changes based on the state read
// and what is in the Component.Spec
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
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

	// Assign the generated ResourceVersion to the resource status
	if instance.Status.RevNumber == "" {
		instance.Status.RevNumber = instance.ObjectMeta.ResourceVersion
	}

	if !instance.ObjectMeta.DeletionTimestamp.IsZero() {
		log.Info("** DELETION **")
		return reconcile.Result{}, nil
	}

	// We only call the pipeline when the component has been created
	// and if the Status Revision Number is the same
	if instance.Status.RevNumber == instance.ObjectMeta.ResourceVersion {
		outputIS, err := r.CreateOutputImageStream(instance)
		if err != nil {
			return reconcile.Result{}, err
		}
		builderIS, err := r.CreateBuilderImageStream(instance)
		if err != nil {
			return reconcile.Result{}, err
		}

		// Get gitsource referenced in component
		gitSource := &gitsourcev1alpha1.GitSource{}
		err = r.client.Get(context.TODO(), client.ObjectKey{
			Namespace: instance.Namespace,
			Name:      instance.Spec.Codebase,
		}, gitSource)
		if err != nil {
			log.Error(err, "Error occured while getting gitsource")
			return reconcile.Result{}, nil
		}

		_, err = r.CreateBuildConfig(instance, builderIS, gitSource)
		if err != nil {
			return reconcile.Result{}, err
		}
		_, err = r.CreateDeploymentConfig(instance, outputIS)
		if err != nil {
			return reconcile.Result{}, err
		}
		log.Info("** ImageStream, BuildConfig and DeploymentConfig successfully created. **")
	}

	return reconcile.Result{}, nil
}

// CreateDeploymentConfig creates a DeploymentConfig OpenShift resource used in S2I.
func (r *ReconcileComponent) CreateDeploymentConfig(cr *componentsv1alpha1.Component, outputIS *imagev1.ImageStream) (*v1.DeploymentConfig, error) {
	dc := newDeploymentConfig(cr, outputIS)
	if err := controllerutil.SetControllerReference(cr, dc, r.scheme); err != nil {
		log.Error(err, "** Setting owner reference fails **")
		return nil, err
	}
	foundDc := &v1.DeploymentConfig{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: dc.Name, Namespace: dc.Namespace}, foundDc)
	if err == nil {
		log.Info("** Skip Creating DeploymentConfig: Already exist", "DeploymentConfig.Namespace", foundDc.Namespace, "DeploymentConfig.Name", foundDc.Name)
		return foundDc, nil
	}
	if errors.IsNotFound(err) {
		log.Info("** Creating a new DeploymentConfig", "DeploymentConfig.Namespace", dc.Namespace, "DeploymentConfig.Name", dc.Name)
		err := r.client.Create(context.TODO(), dc)
		if err != nil {
			log.Error(err, "** DeploymentConfig creation fails **")
			return nil, err
		}
		return dc, nil
	}
	return nil, err
}

// CreateBuildConfig creates a BuildConfig OpenShift resource used in S2I.
func (r *ReconcileComponent) CreateBuildConfig(cr *componentsv1alpha1.Component, builderIS *imagev1.ImageStream, gitSource *gitsourcev1alpha1.GitSource) (*buildv1.BuildConfig, error) {
	bc := r.newBuildConfig(cr, builderIS, gitSource)
	if err := controllerutil.SetControllerReference(cr, bc, r.scheme); err != nil {
		log.Error(err, "** Setting owner reference fails **")
		return nil, err
	}
	foundBc := &buildv1.BuildConfig{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: bc.Name, Namespace: bc.Namespace}, foundBc)
	if err == nil {
		log.Info("** Skip Creating BuildConfig: Already exist", "BuildConfig.Namespace", foundBc.Namespace, "BuildConfig.Name", foundBc.Name)
		return foundBc, nil
	}
	if errors.IsNotFound(err) {
		log.Info("** Creating a new BuildConfig", "BuildConfig.Namespace", bc.Namespace, "BuildConfig.Name", bc.Name)
		err := r.client.Create(context.TODO(), bc)
		if err != nil {
			log.Error(err, "** BuildConfig creation fails **")
			return nil, err
		}
		return bc, nil
	}
	return nil, err
}

// CreateOutputImageStream creates an empty image name that holds the source code of the component to build and deploy.
func (r *ReconcileComponent) CreateOutputImageStream(cr *componentsv1alpha1.Component) (*imagev1.ImageStream, error) {
	outputIS := newOutputImageStream(cr)
	if err := controllerutil.SetControllerReference(cr, outputIS, r.scheme); err != nil {
		log.Error(err, "** Setting owner reference fails **")
		return nil, err
	}

	foundOutputIS := &imagev1.ImageStream{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: outputIS.Name, Namespace: outputIS.Namespace}, foundOutputIS)
	if err == nil {
		log.Info("** Skip Creating output ImageStream: Already exist", "ImageStream.Namespace", foundOutputIS.Namespace, "ImageStream.Name", foundOutputIS.Name)
		return foundOutputIS, nil
	}
	if errors.IsNotFound(err) {
		log.Info("** Creating a new output ImageStream", "ImageStream.Namespace", outputIS.Namespace, "ImageStream.Name", outputIS.Name)
		err := r.client.Create(context.TODO(), outputIS)
		if err != nil {
			log.Error(err, "** output ImageStream creation fails **")
			return nil, err
		}
		return outputIS, nil
	}
	return nil, err
}

// CreateBuilderImageStream either creates an builder image stream fetch from Docker hub or reuse an existing
// image stream in OpenShift namespace.
func (r *ReconcileComponent) CreateBuilderImageStream(instance *componentsv1alpha1.Component) (*imagev1.ImageStream, error) {
	var newImageForBuilder *imagev1.ImageStream
	found := &imagev1.ImageStream{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: instance.Spec.BuildType, Namespace: openshiftNamespace}, found)
	if err == nil {
		log.Info("** Skip Creating builder ImageStream: an OpenShift image already exist", "ImageStream.Namespace", found.Namespace, "ImageStream.Name", found.Name)
		return found, nil
	}
	if errors.IsNotFound(err) { // OpenShift builder image is not present, fallback to create one
		log.Info(fmt.Sprintf("** Searching in namespace %s imagestream %s fails **", openshiftNamespace, instance.Spec.BuildType))
		newImageForBuilder = newImageStreamFromDocker(instance)
		if newImageForBuilder == nil {
			log.Error(err, "** Creating new BUILDER image fails **")
			return nil, errors.NewNotFound(schema.GroupResource{Resource: "ImageStream"}, "builder image for build not found")
		}
		err = r.client.Create(context.TODO(), newImageForBuilder)
		if err != nil {
			log.Error(err, "** Creating new BUILDER image fails **")
			return nil, err
		}
		if err := controllerutil.SetControllerReference(instance, newImageForBuilder, r.scheme); err != nil {
			log.Error(err, "** Setting owner reference fails **")
			return nil, err
		}
	}
	return newImageForBuilder, nil
}
