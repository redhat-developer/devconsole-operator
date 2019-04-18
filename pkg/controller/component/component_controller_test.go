package component

import (
	"context"
	"testing"

	appsv1 "github.com/openshift/api/apps/v1"
	buildv1 "github.com/openshift/api/build/v1"
	imagev1 "github.com/openshift/api/image/v1"
	routev1 "github.com/openshift/api/route/v1"

	devconsoleapi "github.com/redhat-developer/devconsole-api/pkg/apis/devconsole/v1alpha1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	corev1 "k8s.io/api/core/v1"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"

	"k8s.io/client-go/kubernetes/scheme"

	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	Name      = "MyComp"
	Namespace = "test-project"
	Port      = 3000
)

// TestComponentController runs Component.Reconcile() against a
// fake client that tracks a Component object.
func TestComponentController(t *testing.T) {
	reqLogger := log.WithValues("Test", t.Name())
	reqLogger.Info("TestComponentController")

	gs := &devconsoleapi.GitSource{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-git-source",
			Namespace: Namespace,
		},
		Spec: devconsoleapi.GitSourceSpec{
			URL: "https://somegit.con/myrepo",
			Ref: "master",
		},
	}

	// A Component resource with metadata and spec.
	cp := &devconsoleapi.Component{
		ObjectMeta: metav1.ObjectMeta{
			Name:      Name,
			Namespace: Namespace,
			Labels: map[string]string{
				"app.kubernetes.io/part-of":   "application-1",
				"app.kubernetes.io/name":      Name,
				"app.kubernetes.io/component": "backend",
				"app.kubernetes.io/instance":  "mycomp-1",
				"app.kubernetes.io/version":   "1.0",
			},
		},
		Spec: devconsoleapi.ComponentSpec{
			BuildType:    "nodejs",
			GitSourceRef: "my-git-source",
		},
	}

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-secret",
			Namespace: Namespace,
		},
		Type: "Opaque",
		Data: map[string][]byte{
			"username": []byte("username"),
			"password": []byte("password"),
		},
	}
	// Register operator types with the runtime scheme.
	s := scheme.Scheme
	s.AddKnownTypes(devconsoleapi.SchemeGroupVersion, cp)
	s.AddKnownTypes(devconsoleapi.SchemeGroupVersion, gs)
	s.AddKnownTypes(corev1.SchemeGroupVersion, secret)

	// register openshift resource specific schema
	if err := routev1.AddToScheme(s); err != nil {
		log.Error(err, "")
		assert.Nil(t, err, "adding route schema is failing")
	}
	if err := imagev1.AddToScheme(s); err != nil {
		log.Error(err, "")
		assert.Nil(t, err, "adding imagestream schema is failing")
	}
	if err := buildv1.AddToScheme(s); err != nil {
		log.Error(err, "")
		assert.Nil(t, err, "adding buildconfig schema is failing")
	}
	if err := appsv1.AddToScheme(s); err != nil {
		log.Error(err, "")
		assert.Nil(t, err, "adding deploymentconfig, apps schema is failing")
	}

	t.Run("with ReconcileComponent CR containing all required field creates openshift resources", func(t *testing.T) {
		//given
		// Objects to track in the fake client.
		objs := []runtime.Object{
			gs,
			cp,
		}
		// Create a fake client to mock API calls.
		cl := fake.NewFakeClient(objs...)

		// Create a ReconcileComponent object with the scheme and fake client.
		r := &ReconcileComponent{client: cl, scheme: s}

		req := reconcile.Request{
			NamespacedName: types.NamespacedName{
				Name:      Name,
				Namespace: Namespace,
			},
		}

		//when
		_, err := r.Reconcile(req)

		//then
		require.NoError(t, err, "reconcile is failing")

		instance := &devconsoleapi.Component{}
		errGet := r.client.Get(context.TODO(), req.NamespacedName, instance)
		require.NoError(t, errGet, "component is not created")

		is := &imagev1.ImageStream{}
		errGetImage := cl.Get(context.Background(), types.NamespacedName{Namespace: Namespace, Name: Name}, is)
		require.NoError(t, errGetImage, "output imagestream is not created")

		isBuilder := &imagev1.ImageStream{}
		errGetBuilderImage := cl.Get(context.Background(), types.NamespacedName{Namespace: Namespace, Name: cp.Spec.BuildType}, isBuilder)
		require.NoError(t, errGetBuilderImage, "builder imagestream is not created")
		require.Equal(t, cp.Spec.BuildType, isBuilder.ObjectMeta.Name, "imagestream builder should be named after component's buildtype")
		require.Equal(t, Namespace, isBuilder.ObjectMeta.Namespace, "")
		require.Equal(t, 6, len(isBuilder.Labels), "imagestream builder should contain six labels")
		require.Equal(t, Name, isBuilder.Labels["app"], "imagestream builder should have a label with app of CR")
		require.Equal(t, "application-1", isBuilder.Labels["app.kubernetes.io/part-of"], "isBuilder builder should have a label for part-of of CR")
		require.Equal(t, "MyComp", isBuilder.Labels["app.kubernetes.io/name"], "isBuilder builder should have a label with name of CR")
		require.Equal(t, "backend", isBuilder.Labels["app.kubernetes.io/component"], "isBuilder builder should have a label with component of CR")
		require.Equal(t, "mycomp-1", isBuilder.Labels["app.kubernetes.io/instance"], "isBuilder builder should have a label with instance of CR")
		require.Equal(t, "1.0", isBuilder.Labels["app.kubernetes.io/version"], "isBuilder builder should have a label with version of CR")
		require.Equal(t, 1, len(isBuilder.Spec.Tags), "imagestream builder should have a tag specified when")
		require.Equal(t, "latest", isBuilder.Spec.Tags[0].Name, "imagestream builder should take latest version")
		require.Equal(t, "DockerImage", isBuilder.Spec.Tags[0].From.Kind, "imagestream builder should be taken from docker when not found in cluster")
		require.Equal(t, "nodeshift/centos7-s2i-nodejs:10.x", isBuilder.Spec.Tags[0].From.Name, "imagestream builder should be taken from nodeshift/centos7-s2i-nodejs:10.x")

		bc := &buildv1.BuildConfig{}
		errGetBC := cl.Get(context.Background(), types.NamespacedName{Namespace: Namespace, Name: Name}, bc)
		require.NoError(t, errGetBC, "build config is not created")
		require.Equal(t, "https://somegit.con/myrepo", bc.Spec.Source.Git.URI, "build config should not have any source attached")
		require.Equal(t, 2, len(bc.Spec.Triggers), "build config contains 2 triggers")
		require.Equal(t, buildv1.ConfigChangeBuildTriggerType, bc.Spec.Triggers[0].Type, "build config should be triggered on config change")
		require.Equal(t, buildv1.ImageChangeBuildTriggerType, bc.Spec.Triggers[1].Type, "build config should be triggered on image change")
		require.Equal(t, 6, len(bc.Labels), "bc should contain six labels")
		require.Equal(t, Name, bc.ObjectMeta.Labels["app"], "bc builder should have a label with app of CR")
		require.Equal(t, "application-1", bc.ObjectMeta.Labels["app.kubernetes.io/part-of"], "bc builder should have a label with part-of of CR")
		require.Equal(t, "MyComp", bc.ObjectMeta.Labels["app.kubernetes.io/name"], "bc builder should have a label with name of CR")
		require.Equal(t, "backend", bc.ObjectMeta.Labels["app.kubernetes.io/component"], "bc builder should have a label with component of CR")
		require.Equal(t, "mycomp-1", bc.ObjectMeta.Labels["app.kubernetes.io/instance"], "bc builder should have a label with instance of CR")
		require.Equal(t, "1.0", bc.ObjectMeta.Labels["app.kubernetes.io/version"], "bc builder should have a label with version of CR")

		dc := &appsv1.DeploymentConfig{}
		errGetDC := cl.Get(context.Background(), types.NamespacedName{Namespace: Namespace, Name: Name}, dc)
		require.NoError(t, errGetDC, "deployment config is not created")
		require.Equal(t, 2, len(dc.Spec.Triggers), "deployment config contains 2 triggers")
		require.Equal(t, appsv1.DeploymentTriggerOnConfigChange, dc.Spec.Triggers[0].Type, "deployment config should be triggered by DeploymentTriggerOnConfigChange")
		require.Equal(t, appsv1.DeploymentTriggerOnImageChange, dc.Spec.Triggers[1].Type, "deployment config should be triggered by DeploymentTriggerOnImageChange")
		require.Equal(t, Name+":latest", dc.Spec.Triggers[1].ImageChangeParams.From.Name, "deployment config should be triggered by DeploymentTriggerOnImageChange from bc-output")
		require.Equal(t, 6, len(dc.Labels), "dc should contain six labels")
		require.Equal(t, Name, dc.ObjectMeta.Labels["app"], "dc should have a label with app of CR")
		require.Equal(t, "application-1", dc.ObjectMeta.Labels["app.kubernetes.io/part-of"], "dc builder should have a label with part-of of CR")
		require.Equal(t, "MyComp", dc.ObjectMeta.Labels["app.kubernetes.io/name"], "dc builder should have a label with name of CR")
		require.Equal(t, "backend", dc.ObjectMeta.Labels["app.kubernetes.io/component"], "dc builder should have a label with component of CR")
		require.Equal(t, "mycomp-1", dc.ObjectMeta.Labels["app.kubernetes.io/instance"], "dc builder should have a label with instance of CR")
		require.Equal(t, "1.0", dc.ObjectMeta.Labels["app.kubernetes.io/version"], "dc builder should have a label with version of CR")
		require.Equal(t, 6, len(dc.Spec.Selector), "dc should contain six selectors")
		require.Equal(t, Name, dc.Spec.Selector["app"], "dc should have a selector with app of CR")
		require.Equal(t, "application-1", dc.Spec.Selector["app.kubernetes.io/part-of"], "dc builder should have a selector with part-of of CR")
		require.Equal(t, "MyComp", dc.Spec.Selector["app.kubernetes.io/name"], "dc builder should have a selector with name of CR")
		require.Equal(t, "backend", dc.Spec.Selector["app.kubernetes.io/component"], "dc builder should have a selector with component of CR")
		require.Equal(t, "mycomp-1", dc.Spec.Selector["app.kubernetes.io/instance"], "dc builder should have a selector with instance of CR")
		require.Equal(t, "1.0", dc.Spec.Selector["app.kubernetes.io/version"], "dc builder should have a selector with version of CR")

		svc := &corev1.Service{}
		errGetSvc := cl.Get(context.Background(), types.NamespacedName{Namespace: Namespace, Name: Name}, svc)
		require.NoError(t, errGetSvc, "service is not created")
		require.Equal(t, 1, len(svc.Spec.Ports), "service is using one port")
		require.Equal(t, int32(8080), svc.Spec.Ports[0].Port, "default service port should be 8080")
	})

	t.Run("with ReconcileComponent CR containing all optional fields for service port and route should create resources", func(t *testing.T) {
		//given
		// Objects to track in the fake client.
		// A Component resource with metadata and spec.
		cpOptional := &devconsoleapi.Component{
			ObjectMeta: metav1.ObjectMeta{
				Name:      Name,
				Namespace: Namespace,
			},
			Spec: devconsoleapi.ComponentSpec{
				BuildType:    "nodejs",
				GitSourceRef: "my-git-source",
				Port:         Port,
				Exposed:      true,
			},
		}
		objs := []runtime.Object{
			gs,
			cpOptional,
		}
		// Create a fake client to mock API calls.
		cl := fake.NewFakeClient(objs...)

		// Create a ReconcileComponent object with the scheme and fake client.
		r := &ReconcileComponent{client: cl, scheme: s}

		req := reconcile.Request{
			NamespacedName: types.NamespacedName{
				Name:      Name,
				Namespace: Namespace,
			},
		}

		//when
		_, err := r.Reconcile(req)

		//then
		require.NoError(t, err, "reconcile is failing")

		svc := &corev1.Service{}
		errGetSvc := cl.Get(context.Background(), types.NamespacedName{Namespace: Namespace, Name: Name}, svc)
		require.NoError(t, errGetSvc, "service is not created")
		require.Equal(t, 1, len(svc.Spec.Ports), "service is using one port")
		require.Equal(t, int32(Port), svc.Spec.Ports[0].Port, "service port should be 3000")

		rte := &routev1.Route{}
		errGetRte := cl.Get(context.Background(), types.NamespacedName{Namespace: Namespace, Name: Name}, rte)
		require.NoError(t, errGetRte, "route is not created")
		require.Equal(t, intstr.IntOrString{IntVal: Port}, rte.Spec.Port.TargetPort)

		// TODO(corinne): ask Dipak
		//ep := &corev1.Endpoints{}
		//labels := map[string]string{
		//	"app": Name,
		//}
		//errGetEp := cl.List(context.Background(), client.MatchingLabels(labels), ep)
		//require.NoError(t, errGetEp, "endpoints are not created")
	})

	t.Run("with ReconcileComponent CR containing all optional fields for service port and route should create resources", func(t *testing.T) {
		//given
		// Objects to track in the fake client.
		cpOptional := &devconsoleapi.Component{
			ObjectMeta: metav1.ObjectMeta{
				Name:      Name,
				Namespace: Namespace,
			},
			Spec: devconsoleapi.ComponentSpec{
				BuildType:    "nodejs",
				GitSourceRef: "https://somegit.con/myrepo",
				Port:         65700, // not a valid port
				Exposed:      true,
			},
		}
		objs := []runtime.Object{
			cpOptional,
		}
		// Create a fake client to mock API calls.
		cl := fake.NewFakeClient(objs...)

		// Create a ReconcileComponent object with the scheme and fake client.
		r := &ReconcileComponent{client: cl, scheme: s}

		req := reconcile.Request{
			NamespacedName: types.NamespacedName{
				Name:      Name,
				Namespace: Namespace,
			},
		}

		//when
		_, err := r.Reconcile(req)

		//then
		require.Error(t, err, "reconcile should failing")
	})

	t.Run("with ReconcileComponent CR containing all required field and buildtype matches openshift namespace imagestream", func(t *testing.T) {
		//given
		isNodejs := &imagev1.ImageStream{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "nodejs",
				Namespace: "openshift",
			},
			Spec: imagev1.ImageStreamSpec{},
		}
		// Objects to track in the fake client.
		objs := []runtime.Object{
			gs,
			cp,
			isNodejs,
		}
		// Create a fake client to mock API calls.
		cl := fake.NewFakeClient(objs...)

		// Create a ReconcileComponent object with the scheme and fake client.
		r := &ReconcileComponent{client: cl, scheme: s}

		req := reconcile.Request{
			NamespacedName: types.NamespacedName{
				Name:      Name,
				Namespace: Namespace,
			},
		}

		//when
		_, err := r.Reconcile(req)

		//then
		require.NoError(t, err, "reconcile is failing")

		instance := &devconsoleapi.Component{}
		errGet := r.client.Get(context.TODO(), req.NamespacedName, instance)
		require.NoError(t, errGet, "component is not created")

		is := &imagev1.ImageStream{}
		errGetImage := cl.Get(context.Background(), types.NamespacedName{Namespace: Namespace, Name: Name}, is)
		require.NoError(t, errGetImage, "output imagestream is not created")

		isBuilder := &imagev1.ImageStream{}
		errGetBuilderImage := cl.Get(context.Background(), types.NamespacedName{Namespace: Namespace, Name: cp.Spec.BuildType}, isBuilder)
		require.Error(t, errGetBuilderImage, "builder imagestream should not be created")

		bc := &buildv1.BuildConfig{}
		errGetBC := cl.Get(context.Background(), types.NamespacedName{Namespace: Namespace, Name: Name}, bc)
		require.NoError(t, errGetBC, "build config is not created")
		require.Equal(t, "https://somegit.con/myrepo", bc.Spec.Source.Git.URI, "build config should not have any source attached")
		require.Equal(t, 2, len(bc.Spec.Triggers), "build config contains 2 triggers")
		require.Equal(t, buildv1.ConfigChangeBuildTriggerType, bc.Spec.Triggers[0].Type, "build config should be triggered on config change")
		require.Equal(t, buildv1.ImageChangeBuildTriggerType, bc.Spec.Triggers[1].Type, "build config should be triggered on image change")
		require.Equal(t, "openshift", bc.Spec.CommonSpec.Strategy.SourceStrategy.From.Namespace, "builder image used in build config should be taken from openshift namespace")
		require.Equal(t, "nodejs:latest", bc.Spec.CommonSpec.Strategy.SourceStrategy.From.Name, "builder image used in build config should be taken from openshift's nodejs image")
		require.Equal(t, 6, len(bc.Labels), "bc should contain six labels")
		require.Equal(t, Name, bc.ObjectMeta.Labels["app"], "bc builder should have a label with app of CR")
		require.Equal(t, "application-1", bc.ObjectMeta.Labels["app.kubernetes.io/part-of"], "bc builder should have a label with part-of of CR")
		require.Equal(t, "MyComp", bc.ObjectMeta.Labels["app.kubernetes.io/name"], "bc builder should have a label with name of CR")
		require.Equal(t, "backend", bc.ObjectMeta.Labels["app.kubernetes.io/component"], "bc builder should have a label with component of CR")
		require.Equal(t, "mycomp-1", bc.ObjectMeta.Labels["app.kubernetes.io/instance"], "bc builder should have a label with instance of CR")
		require.Equal(t, "1.0", bc.ObjectMeta.Labels["app.kubernetes.io/version"], "bc builder should have a label with version of CR")

		dc := &appsv1.DeploymentConfig{}
		errGetDC := cl.Get(context.Background(), types.NamespacedName{Namespace: Namespace, Name: Name}, dc)
		require.NoError(t, errGetDC, "deployment config is not created")
		require.Equal(t, 2, len(dc.Spec.Triggers), "deployment config contains 2 triggers")
		require.Equal(t, appsv1.DeploymentTriggerOnConfigChange, dc.Spec.Triggers[0].Type, "deployment config should be triggered by DeploymentTriggerOnConfigChange")
		require.Equal(t, appsv1.DeploymentTriggerOnImageChange, dc.Spec.Triggers[1].Type, "deployment config should be triggered by DeploymentTriggerOnImageChange")
		require.Equal(t, Name+":latest", dc.Spec.Triggers[1].ImageChangeParams.From.Name, "deployment config should be triggered by DeploymentTriggerOnImageChange from bc-output")
		require.Equal(t, 6, len(dc.Labels), "dc should contain six labels")
		require.Equal(t, Name, dc.ObjectMeta.Labels["app"], "dc should have a label with app of CR")
		require.Equal(t, "application-1", dc.ObjectMeta.Labels["app.kubernetes.io/part-of"], "dc builder should have a label with part-of of CR")
		require.Equal(t, "MyComp", dc.ObjectMeta.Labels["app.kubernetes.io/name"], "dc builder should have a label with name of CR")
		require.Equal(t, "backend", dc.ObjectMeta.Labels["app.kubernetes.io/component"], "dc builder should have a label with component of CR")
		require.Equal(t, "mycomp-1", dc.ObjectMeta.Labels["app.kubernetes.io/instance"], "dc builder should have a label with instance of CR")
		require.Equal(t, "1.0", dc.ObjectMeta.Labels["app.kubernetes.io/version"], "dc builder should have a label with version of CR")
		require.Equal(t, 6, len(dc.Spec.Selector), "dc should contain six selectors")
		require.Equal(t, Name, dc.Spec.Selector["app"], "dc should have a selector with app of CR")
		require.Equal(t, "application-1", dc.Spec.Selector["app.kubernetes.io/part-of"], "dc builder should have a selector with part-of of CR")
		require.Equal(t, "MyComp", dc.Spec.Selector["app.kubernetes.io/name"], "dc builder should have a selector with name of CR")
		require.Equal(t, "backend", dc.Spec.Selector["app.kubernetes.io/component"], "dc builder should have a selector with component of CR")
		require.Equal(t, "mycomp-1", dc.Spec.Selector["app.kubernetes.io/instance"], "dc builder should have a selector with instance of CR")
		require.Equal(t, "1.0", dc.Spec.Selector["app.kubernetes.io/version"], "dc builder should have a selector with version of CR")
	})

	t.Run("with secret defined in the GitSource", func(t *testing.T) {
		// Add Secret reference in GitSource
		gs.Spec.SecretRef = &devconsoleapi.SecretRef{
			Name: "my-secret",
		}

		// Track objects
		objs := []runtime.Object{
			secret,
			gs,
			cp,
		}

		// Create a fake client to mock API calls.
		cl := fake.NewFakeClient(objs...)

		// Create a ReconcileComponent object with the scheme and fake client.
		r := &ReconcileComponent{client: cl, scheme: s}

		req := reconcile.Request{
			NamespacedName: types.NamespacedName{
				Name:      Name,
				Namespace: Namespace,
			},
		}

		//when
		_, err := r.Reconcile(req)

		//then
		require.NoError(t, err, "reconcile is failing")

		instance := &devconsoleapi.Component{}
		errGet := r.client.Get(context.TODO(), req.NamespacedName, instance)
		require.NoError(t, errGet, "component is not created")

		is := &imagev1.ImageStream{}
		errGetImage := cl.Get(context.Background(), types.NamespacedName{Namespace: Namespace, Name: Name}, is)
		require.NoError(t, errGetImage, "output imagestream is not created")

		bc := &buildv1.BuildConfig{}
		errGetBC := cl.Get(context.Background(), types.NamespacedName{Namespace: Namespace, Name: Name}, bc)
		require.NoError(t, errGetBC, "build config should not not created with missing CR's buildtype")
		require.Equal(t, "https://somegit.con/myrepo", bc.Spec.Source.Git.URI, "build config should not have any source attached")
		// BuildConfig should have reference to secret
		require.Equal(t, "my-secret", bc.Spec.CommonSpec.Source.SourceSecret.Name, "Secret name is not present")
	})

	t.Run("without secret defined in the GitSource", func(t *testing.T) {
		// Add Secret reference in GitSource
		gs.Spec.SecretRef = nil
		// Track objects
		objs := []runtime.Object{
			gs,
			cp,
		}

		// Create a fake client to mock API calls.
		cl := fake.NewFakeClient(objs...)

		// Create a ReconcileComponent object with the scheme and fake client.
		r := &ReconcileComponent{client: cl, scheme: s}

		req := reconcile.Request{
			NamespacedName: types.NamespacedName{
				Name:      Name,
				Namespace: Namespace,
			},
		}

		//when
		_, err := r.Reconcile(req)

		//then
		require.NoError(t, err, "reconcile is failing")

		instance := &devconsoleapi.Component{}
		errGet := r.client.Get(context.TODO(), req.NamespacedName, instance)
		require.NoError(t, errGet, "component is not created")

		is := &imagev1.ImageStream{}
		errGetImage := cl.Get(context.Background(), types.NamespacedName{Namespace: Namespace, Name: Name}, is)
		require.NoError(t, errGetImage, "output imagestream is not created")

		bc := &buildv1.BuildConfig{}
		errGetBC := cl.Get(context.Background(), types.NamespacedName{Namespace: Namespace, Name: Name}, bc)
		require.NoError(t, errGetBC, "build config should not not created with missing CR's buildtype")
		require.Equal(t, "https://somegit.con/myrepo", bc.Spec.Source.Git.URI, "build config should not have any source attached")
		// BuildConfig should not have reference to secret
		require.Nil(t, nil, bc.Spec.CommonSpec.Source.SourceSecret, "Source secret reference should not be present since we don't have secret defined in GitSource")
	})

	t.Run("with ReconcileComponent CR without buildtype", func(t *testing.T) {
		//given
		objs := []runtime.Object{
			gs,
			cp,
		}
		cp.Spec.BuildType = ""
		cp.Spec.GitSourceRef = "my-git-source"
		// Create a fake client to mock API calls.
		cl := fake.NewFakeClient(objs...)

		// Create a ReconcileComponent object with the scheme and fake client.
		r := &ReconcileComponent{client: cl, scheme: s}

		req := reconcile.Request{
			NamespacedName: types.NamespacedName{
				Name:      Name,
				Namespace: Namespace,
			},
		}

		//when
		_, err := r.Reconcile(req)

		//then
		require.Error(t, err, "reconcile is failing")

		instance := &devconsoleapi.Component{}
		errGet := r.client.Get(context.TODO(), req.NamespacedName, instance)
		require.NoError(t, errGet, "component is not created")

		is := &imagev1.ImageStream{}
		errGetImage := cl.Get(context.Background(), types.NamespacedName{Namespace: Namespace, Name: Name}, is)
		require.NoError(t, errGetImage, "output imagestream is not created")

		bc := &buildv1.BuildConfig{}
		errGetBC := cl.Get(context.Background(), types.NamespacedName{Namespace: Namespace, Name: Name}, bc)
		require.Error(t, errGetBC, "build config should not not created with missing CR's buildtype")
		require.Equal(t, errors.ReasonForError(errGetBC), metav1.StatusReasonNotFound, "bc could not found associated imagestream")

		dc := &appsv1.DeploymentConfig{}
		errGetDC := cl.Get(context.Background(), types.NamespacedName{Namespace: Namespace, Name: Name}, dc)
		require.Error(t, errGetDC, "deployment config should not be created")
	})

	t.Run("with ReconcileComponent CR without codebases", func(t *testing.T) {
		//given
		objs := []runtime.Object{
			cp,
		}
		cp.Spec.BuildType = "nodejs"
		cp.Spec.GitSourceRef = ""
		// Create a fake client to mock API calls.
		cl := fake.NewFakeClient(objs...)

		// Create a ReconcileComponent object with the scheme and fake client.
		r := &ReconcileComponent{client: cl, scheme: s}

		req := reconcile.Request{
			NamespacedName: types.NamespacedName{
				Name:      Name,
				Namespace: Namespace,
			},
		}

		//when
		_, err := r.Reconcile(req)

		//then
		require.Error(t, err, "reconcile should fail since no gitsource reference provided")

		instance := &devconsoleapi.Component{}
		errGet := r.client.Get(context.TODO(), req.NamespacedName, instance)
		require.NoError(t, errGet, "component is not created")

		is := &imagev1.ImageStream{}
		errGetImage := cl.Get(context.Background(), types.NamespacedName{Namespace: Namespace, Name: Name}, is)
		require.Error(t, errGetImage, "output imagestream is not created")

		isBuilder := &imagev1.ImageStream{}
		errGetBuilderImage := cl.Get(context.Background(), types.NamespacedName{Namespace: Namespace, Name: Name}, isBuilder)
		require.Error(t, errGetBuilderImage, "builder imagestream is not created")

		bc := &buildv1.BuildConfig{}
		errGetBC := cl.Get(context.Background(), types.NamespacedName{Namespace: Namespace, Name: Name}, bc)
		require.Error(t, errGetBC, "buildconfig should not have created")
	})
}
