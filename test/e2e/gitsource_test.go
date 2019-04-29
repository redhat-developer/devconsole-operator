package e2e

import (
	"context"
	"fmt"
	"os"
	"testing"

	framework "github.com/operator-framework/operator-sdk/pkg/test"
	"github.com/operator-framework/operator-sdk/pkg/test/e2eutil"

	devconsole "github.com/redhat-developer/devconsole-api/pkg/apis"
	devconsoleapi "github.com/redhat-developer/devconsole-api/pkg/apis/devconsole/v1alpha1"

	"github.com/stretchr/testify/require"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// ComponentTest does e2e test as per operator-sdk documentation
// https://github.com/operator-framework/operator-sdk/blob/cc7b175/doc/test-framework/writing-e2e-tests.md
func TestGitsource(t *testing.T) {
	var err error
	gitSourceList := &devconsoleapi.GitSourceList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "GitSource",
			APIVersion: "devconsole.openshift.io/v1alpha1",
		},
	}

	err = framework.AddToFrameworkScheme(devconsole.AddToScheme, gitSourceList)
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
	err = e2eutil.WaitForOperatorDeployment(t, f.KubeClient, os.Getenv("DEPLOYED_NAMESPACE"), "devconsole-operator", 1, retryInterval, timeout)
	require.NoError(t, err, "failed while waiting for operator deployment")

	t.Log("operator is ready and running")

	gs := &devconsoleapi.GitSource{
		TypeMeta: metav1.TypeMeta{
			Kind:       "GitSource",
			APIVersion: "devconsole.openshift.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-git-source-2",
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

	t.Run("retrieve component and verify related resources are created", func(t *testing.T) {
		outputCR := &devconsoleapi.GitSource{}
		err = f.Client.Get(context.TODO(), types.NamespacedName{Name: "my-git-source-2", Namespace: namespace}, outputCR)
		require.NoError(t, err, "failed to retrieve custom resource of kind `GitSource`")
		require.Equal(t, "my-git-source-2", outputCR.ObjectMeta.Name)
		require.Equal(t, namespace, outputCR.ObjectMeta.Namespace)
	})
}
