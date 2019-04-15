package gitsource

import (
	"context"
	"github.com/redhat-developer/devconsole-api/pkg/apis/devconsole/v1alpha1"
	"github.com/redhat-developer/devconsole-git/pkg/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/h2non/gock.v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"testing"
)

const (
	repoIdentifier = "some-org/some-repo"
	repoGitHubURL  = "https://github.com/" + repoIdentifier
)

func TestReconcileGitSourceConnectionOK(t *testing.T) {
	//given
	defer gock.OffAll()
	gs := test.NewGitSource(test.WithURL(repoGitHubURL))
	reconciler, request, client := PrepareClient(test.GitSourceName,
		test.RegisterGvkObject(v1alpha1.SchemeGroupVersion, gs))

	gock.New("https://github.com").
		Get("/some-org/some-repo.git/info/refs").
		MatchParam("service", "git-upload-pack").
		Reply(200).
		BodyString(`004a8d501bc8f3a77129c17a7120bac2d4d70f4d9291 refs/heads/master
003f8c48499a598266ed7ef609070b84d2c8707fb1dd refs/heads/dev
0000`)

	//when
	_, err := reconciler.Reconcile(request)

	//then
	require.NoError(t, err)
	assertGitSource(t, client, v1alpha1.Initializing, v1alpha1.OK, "")
}

func TestReconcileGitSourceConnectionFail(t *testing.T) {
	//given
	gs := test.NewGitSource(test.WithURL(repoGitHubURL))
	gs.Status.State = v1alpha1.Ready
	reconciler, request, client := PrepareClient(test.GitSourceName,
		test.RegisterGvkObject(v1alpha1.SchemeGroupVersion, gs))

	//when
	_, err := reconciler.Reconcile(request)

	//then
	require.NoError(t, err)
	assertGitSource(t, client, v1alpha1.Ready, v1alpha1.Failed, "unable to reach the repo")
}

func TestReconcileGitSourceConnectionSkip(t *testing.T) {
	//given
	gs := test.NewGitSource(test.WithURL(repoGitHubURL))
	gs.Status.Connection.State = v1alpha1.OK
	gs.Status.Connection.Error = "my cool error"
	reconciler, request, client := PrepareClient(test.GitSourceName,
		test.RegisterGvkObject(v1alpha1.SchemeGroupVersion, gs))

	//when
	_, err := reconciler.Reconcile(request)

	//then
	require.NoError(t, err)
	assertGitSource(t, client, "", v1alpha1.OK, "my cool error")
}

func PrepareClient(name string, gvkObjects ...test.GvkObject) (*ReconcileGitSource, reconcile.Request, client.Client) {
	// Create a fake client to mock API calls.
	cl, s := test.PrepareClient(gvkObjects...)

	// Create a ReconcileToolChainEnabler object with the scheme and fake client.
	r := &ReconcileGitSource{client: cl, scheme: s}
	req := test.NewReconcileRequest(name)

	return r, req, cl
}

func assertGitSource(t *testing.T, client client.Client, gsState v1alpha1.State, state v1alpha1.ConnectionState, errorMsg string) {
	gitSource := &v1alpha1.GitSource{}
	err := client.Get(context.TODO(), newNsdName(test.Namespace, test.GitSourceName), gitSource)
	require.NoError(t, err)

	require.NotNil(t, gitSource.Status.Connection)
	if errorMsg != "" {
		assert.Contains(t, gitSource.Status.Connection.Error, errorMsg)
	} else {
		assert.Empty(t, gitSource.Status.Connection.Error)
	}
	assert.NotNil(t, gitSource.Status.Connection)
	assert.Equal(t, state, gitSource.Status.Connection.State)
	assert.Equal(t, gsState, gitSource.Status.State)
}

func newNsdName(namespace, name string) types.NamespacedName {
	return types.NamespacedName{Namespace: namespace, Name: name}
}
