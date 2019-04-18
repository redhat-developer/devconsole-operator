package component

import (
	"context"
	e "errors"
	"fmt"
	v1 "github.com/openshift/api/apps/v1"
	buildv1 "github.com/openshift/api/build/v1"
	imagev1 "github.com/openshift/api/image/v1"
	routev1 "github.com/openshift/api/route/v1"
	devconsoleapi "github.com/redhat-developer/devconsole-api/pkg/apis/devconsole/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
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

var log = logf.Log

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
	err = c.Watch(&source.Kind{Type: &devconsoleapi.Component{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// Watch for changes to secondary resource DeploymentConfig
	err = c.Watch(&source.Kind{Type: &v1.DeploymentConfig{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// Watch for changes to secondary resource BuildConfig
	err = c.Watch(&source.Kind{Type: &buildv1.BuildConfig{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// Watch for changes to secondary resource Service
	err = c.Watch(&source.Kind{Type: &corev1.Service{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// Watch for changes to secondary resource Route
	err = c.Watch(&source.Kind{Type: &routev1.Route{}}, &handler.EnqueueRequestForObject{})
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
	// Fetch the Component instance
	instance := &devconsoleapi.Component{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue.
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request/*  */.
		return reconcile.Result{}, err
	}

	// Checking and logging secondary resource lifecycle
	dcList := &v1.DeploymentConfigList{}
	err = r.ObserveDeploymentConfig(instance, dcList)
	if err != nil {
		return reconcile.Result{}, nil
	}
	bcList := &buildv1.BuildConfigList{}
	err = r.ObserveBuildConfig(instance, bcList)
	if err != nil {
		return reconcile.Result{}, nil
	}

	log.Info("============================================================")
	log.Info(fmt.Sprintf("✨✨ Reconciling Component %s, namespace %s ✨✨", request.Name, request.Namespace))
	log.Info(fmt.Sprintf("** Creation time: %s", instance.ObjectMeta.CreationTimestamp))
	log.Info(fmt.Sprintf("** Resource version: %s", instance.ObjectMeta.ResourceVersion))
	log.Info(fmt.Sprintf("** Generation version: %d", instance.ObjectMeta.Generation))
	log.Info(fmt.Sprintf("** Deletion time: %s", instance.ObjectMeta.DeletionTimestamp))
	log.Info("============================================================")

	// Assign the generated ResourceVersion to the resource status.
	if instance.Status.RevNumber == "" {
		instance.Status.RevNumber = instance.ObjectMeta.ResourceVersion
	}

	if !instance.ObjectMeta.DeletionTimestamp.IsZero() {
		log.Info("👻👻 Deleting component CR 👻👻")
		return reconcile.Result{}, nil
	}

	gitSource, err := r.GetGitSource(instance)
	if err != nil {
		return reconcile.Result{}, err
	}
	outputIS, err := r.CreateOutputImageStream(instance)
	if err != nil {
		return reconcile.Result{}, err
	}
	builderIS, err := r.CreateBuilderImageStream(instance)
	if err != nil {
		return reconcile.Result{}, err
	}
	secret, _ := r.GetSourceSecret(instance, gitSource)
	_, err = r.CreateBuildConfig(instance, builderIS, gitSource, secret)
	if err != nil {
		return reconcile.Result{}, err
	}
	_, err = r.CreateDeploymentConfig(instance, outputIS)
	if err != nil {
		return reconcile.Result{}, err
	}
	_, err = r.CreateService(instance)
	if err != nil {
		return reconcile.Result{}, err
	}
	var route *routev1.Route
	if instance.Spec.Exposed {
		route, err = r.CreateRoute(instance)
		if err != nil {
			return reconcile.Result{}, err
		}
	}
	if instance.Status.RevNumber == instance.ObjectMeta.ResourceVersion {
		log.Info(fmt.Sprintf("🎉🎉  Component %s has been successfully created!  🎉🎉 ", instance.Name))
		if route != nil {
			log.Info(fmt.Sprintf("🎉🎉  Go to http://%s:%d  🎉🎉 ", route.Spec.Host, route.Spec.Port.TargetPort.IntVal))
		}
	}

	return reconcile.Result{}, nil
}

// ObserveBuildConfig watches for secondary resource BuildConfig.
func (r *ReconcileComponent) ObserveBuildConfig(cr *devconsoleapi.Component, bcList *buildv1.BuildConfigList) error {
	lbls := map[string]string{
		"app": cr.Name,
	}
	opts := client.ListOptions{
		Namespace:     cr.Namespace,
		LabelSelector: labels.SelectorFromSet(lbls),
	}
	err := r.client.List(context.TODO(),
		&opts,
		bcList)
	if err != nil {
		log.Error(err, "failed to list existing BuildConfig")
		return err
	}

	for _, bc := range bcList.Items {
		if bc.Status.LastVersion == 0 {
			log.Info(fmt.Sprintf("👻👻  Scaling down BuildConfig %s 👻👻", bc.Name))
			return r.UpdateStatus(cr, devconsoleapi.PhaseBuilding)
		}
	}
	return nil
}

// ObserveDeploymentConfig watches for secondary resource DeploymentConfig.
func (r *ReconcileComponent) ObserveDeploymentConfig(cr *devconsoleapi.Component, dcList *v1.DeploymentConfigList) error {
	lbls := map[string]string{
		"app": cr.Name,
	}
	opts := client.ListOptions{
		Namespace:     cr.Namespace,
		LabelSelector: labels.SelectorFromSet(lbls),
	}
	err := r.client.List(context.TODO(),
		&opts,
		dcList)
	if err != nil {
		log.Error(err, "failed to list existing DeploymentConfig")
		return err
	}

	for _, dc := range dcList.Items {
		if dc.Status.Replicas < dc.Spec.Replicas {
			log.Info(fmt.Sprintf("👻👻  Scaling up DeploymentConfig %s 👻👻", dc.Name))
			return r.UpdateStatus(cr, devconsoleapi.PhaseDeploying)
		} else {
			log.Info(fmt.Sprintf("✨✨ Stable DeploymentConfig %s ✨✨", dc.Name))
			return r.UpdateStatus(cr, devconsoleapi.PhaseDeployed)
		}
	}
	return nil
}

// Update status of component
func (r *ReconcileComponent) UpdateStatus(cr *devconsoleapi.Component, status string) error {
	if cr.Status.Phase != status {
		cr.Status.Phase = status
		err := r.client.Update(context.TODO(), cr)
		if err != nil {
			log.Error(err, "** failed to update component status **")
			return err
		}
	}
	return nil
}

// GetGitSource return the GitSource associated to Component CR.
func (r *ReconcileComponent) GetSourceSecret(cr *devconsoleapi.Component, gitSource *devconsoleapi.GitSource) (*corev1.Secret, error) {
	// Check if secrets provided exist or not
	if gitSource.Spec.SecretRef != nil && gitSource.Spec.SecretRef.Name != "" {
		secret := newSecret(cr, gitSource)
		foundSecret := &corev1.Secret{}
		err := r.client.Get(context.TODO(), types.NamespacedName{Name: secret.Name, Namespace: secret.Namespace}, foundSecret)
		if err == nil {
			log.Info("** Secret found ", "Secret.Namespace", foundSecret.Namespace, "Secret.Name", foundSecret.Name)
			return foundSecret, nil
		}
		if errors.IsNotFound(err) {
			log.Info("** Secret NOT found ", "Secret.Namespace", foundSecret.Namespace, "Secret.Name", foundSecret.Name)
			return nil, err
		}
		return nil, err
	}
	return nil, nil
}

// GetGitSource return the GitSource associated to Component CR.
func (r *ReconcileComponent) GetGitSource(cr *devconsoleapi.Component) (*devconsoleapi.GitSource, error) {
	// Validate if codebase is present since this is mandatory field
	if cr.Spec.GitSourceRef == "" {
		err := e.New("GitSource reference is not provided")
		log.Error(err, "** failed to get gitsource **")
		return nil, err
	}
	// Get gitsource referenced in component
	gitSource := &devconsoleapi.GitSource{}
	err := r.client.Get(context.TODO(), client.ObjectKey{
		Namespace: cr.Namespace,
		Name:      cr.Spec.GitSourceRef,
	}, gitSource)
	if err != nil {
		log.Error(err, "** failed to get gitsource **")
		return nil, err
	}
	return gitSource, nil
}

// CreateRoute creates a route to expose the service if CRD's exposed field is true.
func (r *ReconcileComponent) CreateRoute(cr *devconsoleapi.Component) (*routev1.Route, error) {
	route := newRoute(cr)
	if err := controllerutil.SetControllerReference(cr, route, r.scheme); err != nil {
		log.Error(err, "** Setting owner reference fails **")
		return nil, err
	}
	foundRoute := &routev1.Route{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: route.Name, Namespace: route.Namespace}, foundRoute)
	if err == nil {
		log.Info("** Skip Creating Route: Already exist", "Route.Namespace", foundRoute.Namespace, "Route.Name", foundRoute.Name)
		return foundRoute, nil
	}
	if errors.IsNotFound(err) {
		log.Info("💡💡  Creating a new Route  💡💡", "Route.Namespace", route.Namespace, "Route.Name", route.Name)
		err := r.client.Create(context.TODO(), route)
		if err != nil && !errors.IsAlreadyExists(err) {
			log.Error(err, "** CreateRoute creation fails **")
			return nil, err
		}
		return route, nil
	}
	return nil, err
}

// CreateService creates a service resource to expose the component S2I deployed image.
func (r *ReconcileComponent) CreateService(cr *devconsoleapi.Component) (*corev1.Service, error) {
	port := int32(8080) // default port to 8080
	if cr.Spec.Port != 0 {
		port = cr.Spec.Port
	}
	svc, err := newService(cr, port)
	if err != nil {
		log.Info("** CreateService: Port is not valid")
		return nil, err
	}
	if err := controllerutil.SetControllerReference(cr, svc, r.scheme); err != nil {
		log.Error(err, "** Setting owner reference fails **")
		return nil, err
	}
	foundSvc := &corev1.Service{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: svc.Name, Namespace: svc.Namespace}, foundSvc)
	if err == nil {
		log.Info("** Skip Creating Service: Already exist", "Service.Namespace", foundSvc.Namespace, "Service.Name", foundSvc.Name)
		return foundSvc, nil
	}
	if errors.IsNotFound(err) {
		log.Info("💡💡  Creating a new Service 💡💡", "Service.Namespace", svc.Namespace, "Service.Name", svc.Name)
		err := r.client.Create(context.TODO(), svc)
		if err != nil && !errors.IsAlreadyExists(err) {
			log.Error(err, "** CreateService creation fails **")
			return nil, err
		}
		return svc, nil
	}
	return nil, err
}

// CreateDeploymentConfig creates a DeploymentConfig OpenShift resource used in S2I.
func (r *ReconcileComponent) CreateDeploymentConfig(cr *devconsoleapi.Component, outputIS *imagev1.ImageStream) (*v1.DeploymentConfig, error) {
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
		log.Info("💡💡  Creating a new DeploymentConfig 💡💡", "DeploymentConfig.Namespace", dc.Namespace, "DeploymentConfig.Name", dc.Name)
		err := r.client.Create(context.TODO(), dc)
		if err != nil && !errors.IsAlreadyExists(err) {
			log.Error(err, "** DeploymentConfig creation fails **")
			//r.UpdateStatus(cr, componentsv1alpha1.PhaseError)
			return nil, err
		}
		return dc, nil
	}
	return nil, err
}

// CreateBuildConfig creates a BuildConfig OpenShift resource used in S2I.
func (r *ReconcileComponent) CreateBuildConfig(cr *devconsoleapi.Component, builderIS *imagev1.ImageStream, gitSource *devconsoleapi.GitSource, secret *corev1.Secret) (*buildv1.BuildConfig, error) {
	bc := newBuildConfig(cr, builderIS, gitSource, secret)
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
		log.Info("💡💡 Creating a new BuildConfig 💡💡", "BuildConfig.Namespace", bc.Namespace, "BuildConfig.Name", bc.Name)
		err := r.client.Create(context.TODO(), bc)
		if err != nil && !errors.IsAlreadyExists(err) {
			log.Error(err, "** BuildConfig creation fails **")
			return nil, err
		}
		return bc, nil
	}
	return nil, err
}

// CreateOutputImageStream creates an empty image name that holds the source code of the component to build and deploy.
func (r *ReconcileComponent) CreateOutputImageStream(cr *devconsoleapi.Component) (*imagev1.ImageStream, error) {
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
		log.Info("💡💡  Creating a new output ImageStream 💡💡", "ImageStream.Namespace", outputIS.Namespace, "ImageStream.Name", outputIS.Name)
		err := r.client.Create(context.TODO(), outputIS)
		if err != nil && !errors.IsAlreadyExists(err) {
			log.Error(err, "** output ImageStream creation fails **")
			return nil, err
		}
		return outputIS, nil
	}
	return nil, err
}

// CreateBuilderImageStream either creates an builder image stream fetch from Docker hub or reuse an existing
// image stream in OpenShift namespace.
func (r *ReconcileComponent) CreateBuilderImageStream(cr *devconsoleapi.Component) (*imagev1.ImageStream, error) {
	var newImageForBuilder *imagev1.ImageStream
	found := &imagev1.ImageStream{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: cr.Spec.BuildType, Namespace: openshiftNamespace}, found)
	if err == nil {
		log.Info("** Skip Creating builder ImageStream: an OpenShift image already exist", "ImageStream.Namespace", found.Namespace, "ImageStream.Name", found.Name)
		return found, nil
	}
	if errors.IsNotFound(err) { // OpenShift builder image is not present, fallback to create one.
		log.Info(fmt.Sprintf("** Searching in namespace %s imagestream %s fails **", openshiftNamespace, cr.Spec.BuildType))
		newImageForBuilder = newImageStreamFromDocker(cr)
		if newImageForBuilder == nil {
			log.Error(err, "** Creating new builder image fails **")
			return nil, errors.NewNotFound(schema.GroupResource{Resource: "ImageStream"}, "builder image for build not found")
		}
		log.Info("💡💡  Creating a new builder ImageStream 💡💡")
		err = r.client.Create(context.TODO(), newImageForBuilder)
		if err != nil && !errors.IsAlreadyExists(err) {
			log.Error(err, "** Creating new builder image fails **")
			return nil, err
		}
		if err := controllerutil.SetControllerReference(cr, newImageForBuilder, r.scheme); err != nil {
			log.Error(err, "** Setting owner reference fails **")
			return nil, err
		}
	}
	return newImageForBuilder, nil
}
