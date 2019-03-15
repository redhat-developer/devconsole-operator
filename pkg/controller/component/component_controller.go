package component

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	componentsv1alpha1 "github.com/redhat-developer/devopsconsole-operator/pkg/apis/devopsconsole/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
	"strconv"

	buildv1 "github.com/openshift/api/build/v1"
	imagev1 "github.com/openshift/api/image/v1"
	corev1 "k8s.io/api/core/v1"
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
		// Create an empty image name "myapp-output"
		newImageForOutput := newImageStream(instance.Namespace, instance.Name)
		err = r.client.Create(context.TODO(), newImageForOutput)
		if err != nil {
			log.Error(err, "** Creating new OUTPUT image fails **")
			return reconcile.Result{}, err
		}
		log.Info("** Image stream for OUTPUT created **")
		if err := controllerutil.SetControllerReference(instance, newImageForOutput, r.scheme); err != nil {
			log.Error(err, "** Setting owner reference fails **")
			return reconcile.Result{}, err
		}
		// Create a build image named either "myapp-builder" or reuse openshift's builder image
		ir, err := r.getBuilderImage(instance)
		if err != nil {
			log.Error(err, "** ImageStream builder creation fails **")
			return reconcile.Result{}, err
		}

		// Create build config with s2i
		bc := generateBuildConfig(instance.Namespace, instance.Name, ir, instance.Spec.Codebase, "master")
		err = r.client.Create(context.TODO(), &bc)
		if err != nil {
			log.Error(err, "** BuildConfig creation fails **")
			return reconcile.Result{}, err
		}
		log.Info("** BuildConfig created **")
	}

	return reconcile.Result{}, nil
}

type builderImage struct {
	namespace string
	name      string
}

func (r *ReconcileComponent) getBuilderImage(instance *componentsv1alpha1.Component) (*builderImage, error) {
	var newImageForBuilder *imagev1.ImageStream
	var builderName string
	var builderNamespace string
	// Check if builder image exist in openshift namespace
	found := &imagev1.ImageStream{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: instance.Spec.BuildType, Namespace: openshiftNamespace}, found)
	if err != nil {
		log.Info(fmt.Sprintf("** Searching in namespace %s imagestream %s fails **", openshiftNamespace, instance.Spec.BuildType))
		// Create an empty image name "<BuildType>-builder"
		newImageForBuilder = newImageStreamFromDocker(instance.Namespace, instance.Name, instance.Spec.BuildType)
		if newImageForBuilder == nil {
			log.Error(err, "** Creating new BUILDER image fails **")
			return nil, errors.NewNotFound(schema.GroupResource{Resource: "ImageStream"}, "builder image for build not found")
		}
		err = r.client.Create(context.TODO(), newImageForBuilder)
		if err != nil {
			log.Error(err, "** Creating new BUILDER image fails **")
			return nil, err
		}

		log.Info("** Image stream for BUILDER created **")
		if err := controllerutil.SetControllerReference(instance, newImageForBuilder, r.scheme); err != nil {
			log.Error(err, "** Setting owner reference fails **")
			return nil, err
		}
		builderName = instance.Name + "-builder"
		builderNamespace = instance.Namespace
	} else {
		log.Info("** Found openshift's imagestream to use as builder **")
		newImageForBuilder = found
		builderName = newImageForBuilder.Name
		builderNamespace = newImageForBuilder.Namespace
	}
	return &builderImage{namespace: builderNamespace, name: builderName}, nil
}

func newImageStreamFromDocker(namespace string, name string, buildType string) *imagev1.ImageStream {
	labels := map[string]string{
		"app": name,
	}

	if _, ok := buildTypeImages[buildType]; !ok {

		return nil
	}
	return &imagev1.ImageStream{ObjectMeta: metav1.ObjectMeta{
		Name:      name + "-builder",
		Namespace: namespace,
		Labels:    labels,
	}, Spec: imagev1.ImageStreamSpec{
		LookupPolicy: imagev1.ImageLookupPolicy{
			Local: false,
		},
		Tags: []imagev1.TagReference{
			{
				Name: "latest",
				From: &corev1.ObjectReference{
					Kind: "DockerImage",
					Name: buildTypeImages[buildType],
				},
			},
		},
	}}
}

func newImageStream(namespace string, name string) *imagev1.ImageStream {
	labels := map[string]string{
		"app": name,
	}
	return &imagev1.ImageStream{ObjectMeta: metav1.ObjectMeta{
		Name:      name + "-output",
		Namespace: namespace,
		Labels:    labels,
	}}
}

func getMetaObj(name string, imageNamespace string) metav1.ObjectMeta {
	labels := map[string]string{
		"app": name,
	}
	return metav1.ObjectMeta{Name: name, Namespace: imageNamespace, Labels: labels}
}

func generateBuildConfig(namespace string, name string, builder *builderImage, gitURL string, gitRef string) buildv1.BuildConfig {
	buildSource := buildv1.BuildSource{
		Git: &buildv1.GitBuildSource{
			URI: gitURL,
			Ref: gitRef,
		},
		Type: buildv1.BuildSourceGit,
	}
	incremental := true

	return buildv1.BuildConfig{
		ObjectMeta: getMetaObj(name+"-bc", namespace),
		Spec: buildv1.BuildConfigSpec{
			CommonSpec: buildv1.CommonSpec{
				Output: buildv1.BuildOutput{
					To: &corev1.ObjectReference{
						Kind: "ImageStreamTag",
						Name: name + "-output:latest",
					},
				},
				Source: buildSource,
				Strategy: buildv1.BuildStrategy{
					SourceStrategy: &buildv1.SourceBuildStrategy{
						From: corev1.ObjectReference{
							Kind:      "ImageStreamTag",
							Name:      builder.name + ":latest",
							Namespace: builder.namespace,
						},
						Incremental: &incremental,
					},
				},
			},
			Triggers: []buildv1.BuildTriggerPolicy{
				{
					Type: "ConfigChange",
				}, {
					Type:        "ImageChange",
					ImageChange: &buildv1.ImageChangeTrigger{},
				},
			},
		},
	}
}
