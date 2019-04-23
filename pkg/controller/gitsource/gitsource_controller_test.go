package gitsource

import (
	"context"
	"fmt"
	"github.com/redhat-developer/devconsole-api/pkg/apis/devconsole/v1alpha1"
	"github.com/redhat-developer/devconsole-git/pkg/git/repository"
	"github.com/redhat-developer/devconsole-git/pkg/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/h2non/gock.v1"
	corev1 "k8s.io/api/core/v1"
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
	assertGitSource(t, client, v1alpha1.Ready, v1alpha1.Failed, v1alpha1.RepoNotReachable)
}

func TestReconcileGitSourceConnectionSkip(t *testing.T) {
	//given
	gs := test.NewGitSource(test.WithURL(repoGitHubURL))
	gs.Status.Connection.State = v1alpha1.OK
	gs.Status.Connection.Error = "my cool error"
	gs.Status.Connection.Reason = v1alpha1.ConnectionInternalFailure
	reconciler, request, client := PrepareClient(test.GitSourceName,
		test.RegisterGvkObject(v1alpha1.SchemeGroupVersion, gs))

	//when
	_, err := reconciler.Reconcile(request)

	//then
	require.NoError(t, err)
	assertGitSource(t, client, "", v1alpha1.OK, v1alpha1.ConnectionInternalFailure)
}

func TestValidateGitHubInvalidSecret(t *testing.T) {
	// given
	defer gock.OffAll()
	gock.New("https://api.github.com").
		Get("/user").
		Reply(401)

	secret := test.NewSecret(corev1.SecretTypeOpaque, map[string][]byte{"password": []byte("some-token")})
	gs := test.NewGitSource(test.WithURL(repoGitHubURL))
	gs.Spec.SecretRef = &v1alpha1.SecretRef{Name: test.SecretName}
	reconciler, request, client := PrepareClient(test.GitSourceName,
		test.RegisterGvkObject(v1alpha1.SchemeGroupVersion, gs),
		test.RegisterGvkObject(corev1.SchemeGroupVersion, secret))

	//when
	_, err := reconciler.Reconcile(request)

	// then
	require.NoError(t, err)
	assertGitSource(t, client, v1alpha1.Initializing, v1alpha1.Failed, v1alpha1.BadCredentials)
}

func TestValidateGitLabSecretAndUnavailableRepo(t *testing.T) {
	// given
	defer gock.OffAll()
	gock.New("https://gitlab.com/").
		Get("/api/v4/user").
		Reply(200).
		BodyString("{}")
	gock.New("https://gitlab.com/").
		Get(fmt.Sprintf("/api/v4/projects/%s", repoIdentifier)).
		Reply(404)

	secret := test.NewSecret(corev1.SecretTypeOpaque, map[string][]byte{"password": []byte("some-token")})
	gs := test.NewGitSource(test.WithURL("https://gitlab.com/" + repoIdentifier))
	gs.Spec.SecretRef = &v1alpha1.SecretRef{Name: test.SecretName}
	reconciler, request, client := PrepareClient(test.GitSourceName,
		test.RegisterGvkObject(v1alpha1.SchemeGroupVersion, gs),
		test.RegisterGvkObject(corev1.SchemeGroupVersion, secret))

	//when
	_, err := reconciler.Reconcile(request)

	// then
	require.NoError(t, err)
	assertGitSource(t, client, v1alpha1.Initializing, v1alpha1.Failed, v1alpha1.RepoNotReachable)
}

func TestValidateGitHubSecretAndAvailableRepoWithWrongBranch(t *testing.T) {
	// given
	defer gock.OffAll()
	gock.New("https://api.github.com").
		Get("/user").
		Reply(200)
	gock.New("https://api.github.com").
		Get(fmt.Sprintf("repos/%s", repoIdentifier)).
		Reply(200)
	gock.New("https://api.github.com").
		Get(fmt.Sprintf("repos/%s/branches/any", repoIdentifier)).
		Reply(404)

	secret := test.NewSecret(corev1.SecretTypeOpaque, map[string][]byte{
		"username": []byte("username"),
		"password": []byte("password")})
	gs := test.NewGitSource(test.WithURL(repoGitHubURL), test.WithRef("any"))
	gs.Spec.SecretRef = &v1alpha1.SecretRef{Name: test.SecretName}
	reconciler, request, client := PrepareClient(test.GitSourceName,
		test.RegisterGvkObject(v1alpha1.SchemeGroupVersion, gs),
		test.RegisterGvkObject(corev1.SchemeGroupVersion, secret))

	//when
	_, err := reconciler.Reconcile(request)

	// then
	require.NoError(t, err)
	assertGitSource(t, client, v1alpha1.Initializing, v1alpha1.Failed, v1alpha1.BranchNotFound)
}

func TestValidateBitBucketWithCorrectData(t *testing.T) {
	// given
	defer gock.OffAll()
	gock.New("https://api.bitbucket.org/").
		Get("/2.0/.*").
		Times(3).
		Reply(200)

	secret := test.NewSecret(corev1.SecretTypeOpaque, map[string][]byte{"password": []byte("some-token")})
	gs := test.NewGitSource(test.WithURL("https://bitbucket.org/" + repoIdentifier))
	gs.Spec.SecretRef = &v1alpha1.SecretRef{Name: test.SecretName}
	reconciler, request, client := PrepareClient(test.GitSourceName,
		test.RegisterGvkObject(v1alpha1.SchemeGroupVersion, gs),
		test.RegisterGvkObject(corev1.SchemeGroupVersion, secret))

	// when
	_, err := reconciler.Reconcile(request)

	// then
	require.NoError(t, err)
	assertGitSource(t, client, v1alpha1.Initializing, v1alpha1.OK, "")
}

func TestValidateGenericGit(t *testing.T) {
	//given
	reset := test.RunBasicSshServer(t, "super-secret")
	defer reset()

	dummyRepo := test.NewDummyGitRepo(t, repository.Master)
	dummyRepo.Commit("main.go")
	secret := test.NewSecret(corev1.SecretTypeOpaque, map[string][]byte{"password": []byte("super-secret")})
	gs := test.NewGitSource(test.WithURL("ssh://git@localhost:2222" + dummyRepo.Path))
	gs.Spec.SecretRef = &v1alpha1.SecretRef{Name: test.SecretName}
	reconciler, request, client := PrepareClient(test.GitSourceName,
		test.RegisterGvkObject(v1alpha1.SchemeGroupVersion, gs),
		test.RegisterGvkObject(corev1.SchemeGroupVersion, secret))

	// when
	_, err := reconciler.Reconcile(request)

	// then
	require.NoError(t, err)
	assertGitSource(t, client, v1alpha1.Initializing, v1alpha1.OK, "")
}

func PrepareClient(name string, gvkObjects ...test.GvkObject) (*ReconcileGitSource, reconcile.Request, client.Client) {
	// Create a fake client to mock API calls.
	cl, s := test.PrepareClient(gvkObjects...)

	// Create a ReconcileToolChainEnabler object with the scheme and fake client.
	r := &ReconcileGitSource{client: cl, scheme: s}
	req := test.NewReconcileRequest(name)

	return r, req, cl
}

func assertGitSource(t *testing.T, client client.Client, gsState v1alpha1.State, state v1alpha1.ConnectionState,
	reason v1alpha1.ConnectionFailureReason) {

	gitSource := &v1alpha1.GitSource{}
	err := client.Get(context.TODO(), newNsdName(test.Namespace, test.GitSourceName), gitSource)
	require.NoError(t, err)

	require.NotNil(t, gitSource.Status.Connection)
	if reason != "" {
		assert.Equal(t, gitSource.Status.Connection.Reason, reason)
		assert.NotEmpty(t, gitSource.Status.Connection.Error)
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
