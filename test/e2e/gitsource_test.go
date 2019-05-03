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

// GitSourceTest does e2e test as per operator-sdk documentation
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

	t.Run("check exisiting repo connection should yield OK", func(t *testing.T) {
		gs := &devconsoleapi.GitSource{
			TypeMeta: metav1.TypeMeta{
				Kind:       "GitSource",
				APIVersion: "devconsole.openshift.io/v1alpha1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "my-git-source-1",
				Namespace: namespace,
			},
			Spec: devconsoleapi.GitSourceSpec{
				URL: "https://github.com/fabric8-services/build-tool-detector",
				Ref: "master",
			},
		}
		// use TestCtx's create helper to create the object and add a cleanup function for the new object
		err = f.Client.Create(context.TODO(), gs, &framework.CleanupOptions{TestContext: ctx, Timeout: cleanupTimeout, RetryInterval: cleanupRetryInterval})
		require.NoError(t, err, "failed to create custom resource of kind `GitSource`")

		err = WaitUntilGitSourceReconcile(f, types.NamespacedName{Name: "my-git-source-1", Namespace: namespace})
		if err != nil {
			t.Log("Failed to wait for gitsource reconcile")
			require.NoError(t, err, "Failed to wait for gitsource reconcile")
		}
		outputGS := &devconsoleapi.GitSource{}
		err = f.Client.Get(context.TODO(), types.NamespacedName{Name: "my-git-source-1", Namespace: namespace}, outputGS)
		require.NoError(t, err, "failed to retrieve custom resource of kind `GitSource`")
		require.Equal(t, "my-git-source-1", outputGS.ObjectMeta.Name)
		require.Equal(t, devconsoleapi.OK, outputGS.Status.Connection.State)
		require.Equal(t, namespace, outputGS.ObjectMeta.Namespace)
	})

	t.Run("check non-existing repo connection should yield FAILED", func(t *testing.T) {
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
				URL: "https://example.abc/non-exisiting-repo",
				Ref: "master",
			},
		}
		// use TestCtx's create helper to create the object and add a cleanup function for the new object
		err = f.Client.Create(context.TODO(), gs, &framework.CleanupOptions{TestContext: ctx, Timeout: cleanupTimeout, RetryInterval: cleanupRetryInterval})
		require.NoError(t, err, "failed to create custom resource of kind `GitSource`")

		err = WaitUntilGitSourceReconcile(f, types.NamespacedName{Name: "my-git-source-2", Namespace: namespace})
		if err != nil {
			t.Log("Failed to wait for gitsource reconcile ", err)
			require.NoError(t, err, "Failed to wait for gitsource reconcile")
		}
		outputGS := &devconsoleapi.GitSource{}
		err = f.Client.Get(context.TODO(), types.NamespacedName{Name: "my-git-source-2", Namespace: namespace}, outputGS)
		require.NoError(t, err, "failed to retrieve custom resource of kind `GitSource`")
		require.Equal(t, "my-git-source-2", outputGS.ObjectMeta.Name)
		require.Equal(t, devconsoleapi.Failed, outputGS.Status.Connection.State)
		require.Equal(t, namespace, outputGS.ObjectMeta.Namespace)
	})
}
