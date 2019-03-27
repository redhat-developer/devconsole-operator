package olm

import (
	"context"
	"fmt"
	"testing"
	"time"

	framework "github.com/operator-framework/operator-sdk/pkg/test"
	"github.com/operator-framework/operator-sdk/pkg/test/e2eutil"

	"github.com/redhat-developer/devopsconsole-operator/pkg/apis"
	componentsv1alpha1 "github.com/redhat-developer/devopsconsole-operator/pkg/apis/devopsconsole/v1alpha1"

	"github.com/stretchr/testify/require"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

const (
	retryInterval        = time.Second * 5
	timeout              = time.Second * 60
	cleanupRetryInterval = time.Second * 1
	cleanupTimeout       = time.Second * 5
)

// ComponentTest does e2e test as per operator-sdk documentation
// https://github.com/operator-framework/operator-sdk/blob/cc7b175/doc/test-framework/writing-e2e-tests.md
func TestComponent(t *testing.T) {
	var err error
	// Register types with framework scheme
	componentList := &componentsv1alpha1.ComponentList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Component",
			APIVersion: "devopsconsole.openshift.io/v1alpha1",
		},
	}
	err = framework.AddToFrameworkScheme(apis.AddToScheme, componentList)
	if err != nil {
		t.Fatalf("failed to add custom resource scheme to framework: %v", err)
	}

	ctx := framework.NewTestCtx(t)
	defer ctx.Cleanup()
	err = ctx.InitializeClusterResources(&framework.CleanupOptions{TestContext: ctx, Timeout: cleanupTimeout, RetryInterval: cleanupRetryInterval})
	require.NoError(t, err, "failed to initialize cluster resources")
	t.Log("Initialized cluster resources")

	namespace, err := ctx.GetNamespace()
	require.NoError(t, err, "failed to get namespace where operator needs to run")

	// get global framework variables
	f := framework.Global
	t.Log(fmt.Sprintf("namespace: %s", namespace))
	// wait for component-operator to be ready
	err = e2eutil.WaitForDeployment(t, f.KubeClient, "operators", "devopsconsole-operator", 1, retryInterval, timeout*2)
	require.NoError(t, err, "failed while waiting for operator deployment")

	t.Log("component-operator is ready and running state")

	// create a Component custom resource
	cr := &componentsv1alpha1.Component{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Component",
			APIVersion: "devopsconsole.openshift.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "mycomp",
			Namespace: namespace,
		},
		Spec: componentsv1alpha1.ComponentSpec{
			BuildType: "nodejs",
			Codebase:  "https://github.com/nodeshift-starters/nodejs-rest-http-crud",
		},
	}
	// use TestCtx's create helper to create the object and add a cleanup function for the new object
	err = f.Client.Create(context.TODO(), cr, &framework.CleanupOptions{TestContext: ctx, Timeout: cleanupTimeout, RetryInterval: cleanupRetryInterval})
	require.NoError(t, err, "failed to create custom resource of kind `Component`")

	t.Run("retrieve component and verify related resources are created", func(t *testing.T) {
		cr2 := &componentsv1alpha1.Component{}
		err = f.Client.Get(context.TODO(), types.NamespacedName{Name: "mycomp", Namespace: namespace}, cr2)
		require.NoError(t, err, "failed to retrieve custom resource of kind `Component`")
		// FIXME: These assertions not working
		//require.Equal(t, "Component", cr2.TypeMeta.Kind)
		//require.Equal(t, "devopsconsole.openshift.io/v1alpha1", cr2.TypeMeta.APIVersion)
		require.Equal(t, "mycomp", cr2.ObjectMeta.Name)
		require.Equal(t, namespace, cr2.ObjectMeta.Namespace)
		require.Equal(t, "https://github.com/nodeshift-starters/nodejs-rest-http-crud", cr2.Spec.Codebase)
		require.Equal(t, "nodejs", cr2.Spec.BuildType)
	})
}
