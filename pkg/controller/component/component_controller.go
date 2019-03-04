package component

import (
	"context"
	"fmt"
	tmpl "github.com/redhat-developer/devopsconsole-operator/pkg/util/template"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/util/yaml"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"strconv"
	"strings"
	"text/template"

	componentsv1alpha1 "github.com/redhat-developer/devopsconsole-operator/pkg/apis/components/v1alpha1"
	deploymentconfig "github.com/openshift/api/apps/v1"
	build "github.com/openshift/api/build/v1"
	image "github.com/openshift/api/image/v1"
	route "github.com/openshift/api/route/v1"
	cgoscheme "k8s.io/client-go/kubernetes/scheme"

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

var (
	scheme      = runtime.NewScheme()
	codecs      = serializer.NewCodecFactory(scheme)
	decoderFunc = decoder
	log 		= logf.Log.WithName("controller_component")
)
func init() {
	// Add the standard kubernetes [GVK:Types] type registry
	// e.g (v1,Pods):&v1.Pod{}
	v1.AddToGroupVersion(scheme, schema.GroupVersion{Version: "v1"})
	cgoscheme.AddToScheme(scheme)

	//add openshift types
	deploymentconfig.AddToScheme(scheme)
	image.AddToScheme(scheme)
	route.AddToScheme(scheme)
	build.AddToScheme(scheme)
}
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

	templateObj, ok := tmpl.Templates["outerloop/imagestream"]
	if ok {
		err := CreateResource(templateObj, instance, r.client, r.scheme)
		if err != nil {
			return reconcile.Result{}, err
		}
		log.Info("***** Created Imagestream used as target image to run the application")
	}

	templateObj, ok = tmpl.Templates["outerloop/buildconfig"]
	if ok {
		err := CreateResource(templateObj, instance, r.client, r.scheme)
		if err != nil {
			return reconcile.Result{}, err
		}
		log.Info("***** Created Buildconfig")
	}

	return reconcile.Result{}, nil
}

// TODO move the code outside controller into util

func CreateResource(templateObj template.Template, component *componentsv1alpha1.Component, c client.Client, scheme *runtime.Scheme) error {
	res, err := newResourceFromTemplate(templateObj, component, scheme)
	if err != nil {
		return err
	}

	for _, r := range res {
		if obj, ok := r.(v1.Object); ok {
			obj.SetLabels(PopulateK8sLabels(component))
		}
		err = c.Create(context.TODO(), r)
		if err != nil && errors.IsNotFound(err) {
			return err
		}
	}

	return nil
}

func newResourceFromTemplate(templateObj template.Template, component *componentsv1alpha1.Component, scheme *runtime.Scheme) ([]runtime.Object, error) {
	var result = []runtime.Object{}

	var b = tmpl.Parse(templateObj, component)
	r, err := PopulateKubernetesObjectFromYaml(b.String())
	if err != nil {
		return nil, err
	}

	if strings.HasSuffix(r.GetKind(), "List") {
		l, err := r.ToList()
		if err != nil {
			return nil, err
		}
		for _, item := range l.Items {
			obj, err := RuntimeObjectFromUnstructured(&item)
			if err != nil {
				return nil, err
			}
			ro, ok := obj.(v1.Object)
			ro.SetNamespace(component.Namespace)
			if !ok {
				return nil, err
			}
			controllerutil.SetControllerReference(component, ro, scheme)
			//kubernetes.SetNamespaceAndOwnerReference(obj, component)
			result = append(result, obj)
		}
	} else {
		obj, err := RuntimeObjectFromUnstructured(r)
		if err != nil {
			return nil, err
		}

		ro, ok := obj.(v1.Object)
		ro.SetNamespace(component.Namespace)
		if !ok {
			return nil, err
		}
		controllerutil.SetControllerReference(component, ro, scheme)
		//kubernetes.SetNamespaceAndOwnerReference(obj, component)
		result = append(result, obj)
	}
	return result, nil
}

func PopulateKubernetesObjectFromYaml(data string) (*unstructured.Unstructured, error) {
	yml := []byte(data)
	json, err := yaml.ToJSON(yml)
	if err != nil {
		return nil, err
	}
	u := unstructured.Unstructured{}
	err = u.UnmarshalJSON(json)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// RuntimeObjectFromUnstructured converts an unstructured to a runtime object
func RuntimeObjectFromUnstructured(u *unstructured.Unstructured) (runtime.Object, error) {
	gvk := u.GroupVersionKind()
	decoder := decoder(gvk.GroupVersion(), codecs)

	b, err := u.MarshalJSON()
	if err != nil {
		return nil, fmt.Errorf("error running MarshalJSON on unstructured object: %v", err)
	}
	ro, _, err := decoder.Decode(b, &gvk, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decode json data with gvk(%v): %v", gvk.String(), err)
	}
	return ro, nil
}

func PopulateK8sLabels(component *componentsv1alpha1.Component) map[string]string {
	labels := map[string]string{}
	//labels[componentsv1alpha1.RuntimeLabelKey] = component.Spec.Runtime
	//labels[componentsv1alpha1.RuntimeVersionLabelKey] = component.Spec.Version
	//labels[componentsv1alpha1.ComponentLabelKey] = componentType
	labels["app.kubernetes.io/name"] = component.Name
	//labels[componentsv1alpha1.ManagedByLabelKey] = "component-operator"
	return labels
}

func decoder(gv schema.GroupVersion, codecs serializer.CodecFactory) runtime.Decoder {
	codec := codecs.UniversalDecoder(gv)
	return codec
}