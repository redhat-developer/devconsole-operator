package component

import (
	"context"
	buildv1 "github.com/openshift/api/build/v1"
	imagev1 "github.com/openshift/api/image/v1"
	compv1alpha1 "github.com/redhat-developer/devopsconsole-operator/pkg/apis/devopsconsole/v1alpha1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"testing"
)
const (
	Name = "MyComp"
	Namespace = "test-project"
)
// TestComponentController runs Component.Reconcile() against a
// fake client that tracks a Component object.
func TestComponentController(t *testing.T) {
	reqLogger := log.WithValues("Test", t.Name())
	reqLogger.Info("TestComponentController")

	// A Component resource with metadata and spec.
	cp := &compv1alpha1.Component{
		ObjectMeta: metav1.ObjectMeta{
			Name: Name,
			Namespace: Namespace,
	    },
		Spec: compv1alpha1.ComponentSpec{
			BuildType: "nodejs",
			Codebase: "https://somegit.con/myrepo",
		},
	}

	// Register operator types with the runtime scheme.
	s := scheme.Scheme
	s.AddKnownTypes(compv1alpha1.SchemeGroupVersion, cp)

	// register openshift resource specific schema
	if err := imagev1.AddToScheme(s); err != nil {
		log.Error(err, "")
		assert.Nil(t, err, "adding imagestream schema is failing")
	}
	if err := buildv1.AddToScheme(s); err != nil {
		log.Error(err, "")
		assert.Nil(t, err, "adding buildconfig schema is failing")
	}

	t.Run("with ReconcileComponent CR containing all required field creates openshift resources", func(t *testing.T) {
		//given
		// Objects to track in the fake client.
		objs := []runtime.Object{
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

		instance := &compv1alpha1.Component{}
		errGet := r.client.Get(context.TODO(), req.NamespacedName, instance)
		require.NoError(t, errGet, "component is not created")

		is := &imagev1.ImageStream{}
		errGetImage := cl.Get(context.Background(), types.NamespacedName{Namespace: Namespace, Name: Name + "-output"}, is)
		require.NoError(t, errGetImage, "output imagestream is not created")

		isBuilder := &imagev1.ImageStream{}
		errGetBuilderImage := cl.Get(context.Background(), types.NamespacedName{Namespace: Namespace, Name: Name + "-builder"}, isBuilder)
		require.NoError(t, errGetBuilderImage, "builder imagestream is not created")
		require.Equal(t, isBuilder.ObjectMeta.Name, Name + "-builder", "imagestream builder shoulbe name with pattern CR's name append with -builder")
		require.Equal(t, isBuilder.ObjectMeta.Namespace, Namespace, "")
		require.Equal(t, len(isBuilder.Labels), 1, "imagestream builder should contain one label")
		require.Equal(t, isBuilder.Labels["app"], Name, "imagestream builder should have one label with name of CR.")
		require.Equal(t, len(isBuilder.Spec.Tags), 1, "imagestream builder should have a tag specified when")
		require.Equal(t, isBuilder.Spec.Tags[0].Name, "latest", "imagestream builder should take latest version")
		require.Equal(t, isBuilder.Spec.Tags[0].From.Kind, "DockerImage", "imagestream builder should be taken from docker when not found in cluster")
		require.Equal(t, isBuilder.Spec.Tags[0].From.Name, "nodeshift/centos7-s2i-nodejs:10.x", "imagestream builder should be taken from nodeshift/centos7-s2i-nodejs:10.x")

		bc := &buildv1.BuildConfig{}
		errGetBC := cl.Get(context.Background(), types.NamespacedName{Namespace:Namespace, Name: Name  + "-bc"}, bc)
		require.NoError(t, errGetBC, "build config is not created")
		require.Equal(t, "https://somegit.con/myrepo", bc.Spec.Source.Git.URI, "build config should not have any source attached")
		require.Equal(t, 2, len(bc.Spec.Triggers), "build config contains 2 triggers")
		require.Equal(t, buildv1.ConfigChangeBuildTriggerType, bc.Spec.Triggers[0].Type, "")
		require.Equal(t, buildv1.ImageChangeBuildTriggerType, bc.Spec.Triggers[1].Type, "")
		require.Equal(t, 1, len(bc.Labels), "bc should contain one label")
		require.Equal(t, Name + "-bc", bc.ObjectMeta.Labels["app"], "bc builder should have one label with name of CR.")
	})

	t.Run("with ReconcileComponent CR without buildtype", func(t *testing.T) {
		//given
		objs := []runtime.Object{
			cp,
		}
		cp.Spec.BuildType = ""
		cp.Spec.Codebase = "https://somegit.con/myrepo"
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

		instance := &compv1alpha1.Component{}
		errGet := r.client.Get(context.TODO(), req.NamespacedName, instance)
		require.NoError(t, errGet, "component is not created")

		is := &imagev1.ImageStream{}
		errGetImage := cl.Get(context.Background(), types.NamespacedName{Namespace: Namespace, Name: Name + "-output"}, is)
		require.NoError(t, errGetImage, "output imagestream is not created")

		isBuilder := &imagev1.ImageStream{}
		errGetBuilderImage := cl.Get(context.Background(), types.NamespacedName{Namespace: Namespace, Name: Name + "-builder"}, isBuilder)
		require.Error(t, errGetBuilderImage, "builder imagestream should not be created with missing CR's buildtype")
		require.Equal(t, errors.ReasonForError(errGetBuilderImage), metav1.StatusReasonNotFound, "bc could not found associated imagestream ")

		bc := &buildv1.BuildConfig{}
		errGetBC := cl.Get(context.Background(), types.NamespacedName{Namespace:Namespace, Name: Name  + "-bc"}, bc)
		require.Error(t, errGetBC, "build config should not not created with missing CR's buildtype")
		require.Equal(t, errors.ReasonForError(errGetBC), metav1.StatusReasonNotFound, "bc could not found associated imagestream ")
	})

	t.Run("with ReconcileComponent CR without codebases", func(t *testing.T) {
		//given
		objs := []runtime.Object{
			cp,
		}
		cp.Spec.BuildType = "nodejs"
		cp.Spec.Codebase = ""
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

		instance := &compv1alpha1.Component{}
		errGet := r.client.Get(context.TODO(), req.NamespacedName, instance)
		require.NoError(t, errGet, "component is not created")

		is := &imagev1.ImageStream{}
		errGetImage := cl.Get(context.Background(), types.NamespacedName{Namespace: Namespace, Name: Name + "-output"}, is)
		require.NoError(t, errGetImage, "output imagestream is not created")

		isBuilder := &imagev1.ImageStream{}
		errGetBuilderImage := cl.Get(context.Background(), types.NamespacedName{Namespace: Namespace, Name: Name + "-builder"}, isBuilder)
		require.NoError(t, errGetBuilderImage, "builder imagestream is not created")

		bc := &buildv1.BuildConfig{}
		errGetBC := cl.Get(context.Background(), types.NamespacedName{Namespace:Namespace, Name: Name  + "-bc"}, bc)
		require.NoError(t, errGetBC, "buildconfig is not created")
		require.Equal(t, "", bc.Spec.Source.Git.URI, "build config should not have any source attached")
	})
}