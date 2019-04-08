package e2e

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	framework "github.com/operator-framework/operator-sdk/pkg/test"
	"github.com/operator-framework/operator-sdk/pkg/test/e2eutil"

	devconsole "github.com/redhat-developer/devconsole-api/pkg/apis"
	devconsoleapi "github.com/redhat-developer/devconsole-api/pkg/apis/devconsole/v1alpha1"
	"github.com/redhat-developer/devconsole-operator/pkg/apis"
	componentsv1alpha1 "github.com/redhat-developer/devconsole-operator/pkg/apis/devconsole/v1alpha1"

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
	gitSourceList := &devconsoleapi.GitSourceList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "GitSource",
			APIVersion: "devconsole.openshift.io/v1alpha1",
		},
	}
	// Register types with framework scheme
	componentList := &componentsv1alpha1.ComponentList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Component",
			APIVersion: "devconsole.openshift.io/v1alpha1",
		},
	}

	err = framework.AddToFrameworkScheme(devconsole.AddToScheme, gitSourceList)
	if err != nil {
		t.Fatalf("failed to add custom resource scheme to framework: %v", err)
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
	err = e2eutil.WaitForDeployment(t, f.KubeClient, os.Getenv("DEPLOYED_NAMESPACE"), "devconsole-operator", 1, retryInterval, timeout*2)
	require.NoError(t, err, "failed while waiting for operator deployment")

	t.Log("component is ready and running")

	gs := &devconsoleapi.GitSource{
		TypeMeta: metav1.TypeMeta{
			Kind:       "GitSource",
			APIVersion: "devconsole.openshift.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-git-source",
			Namespace: namespace,
		},
		Spec: devconsoleapi.GitSourceSpec{
			URL: "https://somegit.con/myrepo",
			Ref: "master",
		},
	}
	// use TestCtx's create helper to create the object and add a cleanup function for the new object
	err = f.Client.Create(context.TODO(), gs, &framework.CleanupOptions{TestContext: ctx, Timeout: cleanupTimeout, RetryInterval: cleanupRetryInterval})
	require.NoError(t, err, "failed to create custom resource of kind `GitSource`")

	// create a Component custom resource
	inputCR := &componentsv1alpha1.Component{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Component",
			APIVersion: "devconsole.openshift.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "mycomp",
			Namespace: namespace,
		},
		Spec: componentsv1alpha1.ComponentSpec{
			BuildType:    "nodejs",
			GitSourceRef: "my-git-source",
		},
	}
	// use TestCtx's create helper to create the object and add a cleanup function for the new object
	err = f.Client.Create(context.TODO(), inputCR, &framework.CleanupOptions{TestContext: ctx, Timeout: cleanupTimeout, RetryInterval: cleanupRetryInterval})
	require.NoError(t, err, "failed to create custom resource of kind `Component`")

	t.Run("retrieve component and verify related resources are created", func(t *testing.T) {
		outputCR := &componentsv1alpha1.Component{}
		err = f.Client.Get(context.TODO(), types.NamespacedName{Name: "mycomp", Namespace: namespace}, outputCR)
		require.NoError(t, err, "failed to retrieve custom resource of kind `Component`")
		// FIXME: Uncomment these lines after upgrading dependency versions
		// The following (2) statements will fail due to
		// https://github.com/kubernetes-sigs/controller-runtime/issues/202
		// This issue is resolved in controller-runtime 0.1.8
		//require.Equal(t, "Component", cr2.TypeMeta.Kind)
		//require.Equal(t, "devconsole.openshift.io/v1alpha1", cr2.TypeMeta.APIVersion)
		require.Equal(t, "mycomp", outputCR.ObjectMeta.Name)
		require.Equal(t, namespace, outputCR.ObjectMeta.Namespace)
		require.Equal(t, "my-git-source", outputCR.Spec.GitSourceRef)
		require.Equal(t, "nodejs", outputCR.Spec.BuildType)
		require.Equal(t, "", outputCR.Status.RevNumber)
	})
}
