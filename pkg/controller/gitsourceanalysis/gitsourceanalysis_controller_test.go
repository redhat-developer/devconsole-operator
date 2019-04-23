package gitsourceanalysis

import (
	"context"
	"encoding/base64"
	"fmt"
	"testing"

	"github.com/redhat-developer/devconsole-api/pkg/apis/devconsole/v1alpha1"
	"github.com/redhat-developer/devconsole-git/pkg/git/detector/build"
	"github.com/redhat-developer/devconsole-git/pkg/git/repository"
	"github.com/redhat-developer/devconsole-git/pkg/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/h2non/gock.v1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	pathToTestDir  = "../../../vendor/github.com/redhat-developer/devconsole-git/pkg/test"
	repoIdentifier = "some-org/some-repo"
	repoGitHubURL  = "https://github.com/" + repoIdentifier
)

func TestReconcileGitSourceAnalysisFromGitHubWithDefaultCredentials(t *testing.T) {
	//given
	defer gock.OffAll()
	gs := test.NewGitSource(test.WithURL(repoGitHubURL))
	gsa := test.NewGitSourceAnalysis(test.GitSourceName)
	reconciler, request, client := PrepareClient(test.GitSourceAnalysisName,
		test.RegisterGvkObject(v1alpha1.SchemeGroupVersion, gs, gsa))
	test.MockGHHeadCalls(repoIdentifier, "master", test.S("pom.xml", "mvnw"), matchBasicAuth("anonymous:"))

	//when
	_, err := reconciler.Reconcile(request)

	//then
	require.NoError(t, err)

	assertGitSourceAnalysis(t, client, "", test.S(), buildType(build.Maven, "pom.xml"))
}

func TestReconcileGitSourceAnalysisFromGitHubWithGivenBasicCredentials(t *testing.T) {
	//given
	defer gock.OffAll()
	for _, secretType := range []corev1.SecretType{corev1.SecretTypeOpaque, corev1.SecretTypeBasicAuth} {
		secret := test.NewSecret(secretType, map[string][]byte{
			"username": []byte("username"),
			"password": []byte("password")})

		gs := test.NewGitSource(test.WithURL(repoGitHubURL))
		gs.Spec.SecretRef = &v1alpha1.SecretRef{Name: test.SecretName}
		gsa := test.NewGitSourceAnalysis(test.GitSourceName)
		reconciler, request, client := PrepareClient(test.GitSourceAnalysisName,
			test.RegisterGvkObject(v1alpha1.SchemeGroupVersion, gs, gsa),
			test.RegisterGvkObject(corev1.SchemeGroupVersion, secret))
		langs := test.S("Ruby", "Java", "Go")
		test.MockGHGetApiCalls(t, repoIdentifier, "master", test.S("pom.xml", "main.go"), langs,
			matchBasicAuth("username:password"))

		//when
		_, err := reconciler.Reconcile(request)

		//then
		require.NoError(t, err)
		assertGitSourceAnalysis(t, client, "", langs,
			buildType(build.Maven, "pom.xml"), buildType(build.Golang, "main.go"))
	}
}

func TestReconcileGitSourceAnalysisFromGitHubWithWrongURL(t *testing.T) {
	//given
	defer gock.OffAll()

	gs := test.NewGitSource(test.WithURL(repoGitHubURL))
	gsa := test.NewGitSourceAnalysis(test.GitSourceName)
	reconciler, request, client := PrepareClient(test.GitSourceAnalysisName,
		test.RegisterGvkObject(v1alpha1.SchemeGroupVersion, gs, gsa))
	test.MockNotFoundGitHub(repoIdentifier)

	//when
	_, err := reconciler.Reconcile(request)

	//then
	require.NoError(t, err)
	assertGitSourceAnalysis(t, client, "", nil)
}

func TestReconcileGitSourceAnalysisFromGitHubWithWrongSecret(t *testing.T) {
	//given
	defer gock.OffAll()
	secret := test.NewSecret(corev1.SecretTypeTLS, map[string][]byte{"tls.crt": []byte("crt")})

	gs := test.NewGitSource(test.WithURL(repoGitHubURL))
	gs.Spec.SecretRef = &v1alpha1.SecretRef{Name: test.SecretName}
	gsa := test.NewGitSourceAnalysis(test.GitSourceName)
	reconciler, request, client := PrepareClient(test.GitSourceAnalysisName,
		test.RegisterGvkObject(v1alpha1.SchemeGroupVersion, gs, gsa),
		test.RegisterGvkObject(corev1.SchemeGroupVersion, secret))
	//when
	_, err := reconciler.Reconcile(request)

	//then
	require.NoError(t, err)
	assertGitSourceAnalysis(t, client,
		"the provided secret does not contain any of the required parameters: [username,password,ssh-privatekey] or they are empty", nil)
}

func TestReconcileGitSourceAnalysisWithDifferentGitSourceName(t *testing.T) {
	//given
	defer gock.OffAll()

	gs := test.NewGitSource(test.WithURL(repoGitHubURL))
	gsa := test.NewGitSourceAnalysis(test.GitSourceName)
	gsa.Spec.GitSourceRef.Name = "some-other-git-source"
	reconciler, request, client := PrepareClient(test.GitSourceAnalysisName,
		test.RegisterGvkObject(v1alpha1.SchemeGroupVersion, gs, gsa))
	//when
	_, err := reconciler.Reconcile(request)

	//then
	require.NoError(t, err)
	assertGitSourceAnalysis(t, client, "Failed to fetch the input source", nil)
}

func TestReconcileGitSourceAnalysisWithDifferentSecretName(t *testing.T) {
	//given
	defer gock.OffAll()
	secret := test.NewSecret(corev1.SecretTypeOpaque, map[string][]byte{"password": []byte("some-token")})

	gs := test.NewGitSource(test.WithURL(repoGitHubURL))
	gs.Spec.SecretRef = &v1alpha1.SecretRef{Name: "some-other-secret"}
	gsa := test.NewGitSourceAnalysis(test.GitSourceName)
	reconciler, request, client := PrepareClient(test.GitSourceAnalysisName,
		test.RegisterGvkObject(v1alpha1.SchemeGroupVersion, gs, gsa),
		test.RegisterGvkObject(corev1.SchemeGroupVersion, secret))
	//when
	_, err := reconciler.Reconcile(request)

	//then
	require.NoError(t, err)
	assertGitSourceAnalysis(t, client, "failed to fetch the secret object", nil)
}

func TestReconcileGitSourceAnalysisFromGitHubWithGivenToken(t *testing.T) {
	//given
	defer gock.OffAll()
	for _, secretType := range []corev1.SecretType{corev1.SecretTypeOpaque, corev1.SecretTypeBasicAuth} {
		secret := test.NewSecret(secretType, map[string][]byte{"password": []byte("some-token")})

		gs := test.NewGitSource(test.WithURL(repoGitHubURL))
		gs.Spec.SecretRef = &v1alpha1.SecretRef{Name: test.SecretName}
		gsa := test.NewGitSourceAnalysis(test.GitSourceName)
		reconciler, request, client := PrepareClient(test.GitSourceAnalysisName,
			test.RegisterGvkObject(v1alpha1.SchemeGroupVersion, gs, gsa),
			test.RegisterGvkObject(corev1.SchemeGroupVersion, secret))
		langs := test.S("Ruby")
		test.MockGHGetApiCalls(t, repoIdentifier, "master", test.S("Gemfile", "any"), langs,
			matchToken("some-token"))

		//when
		_, err := reconciler.Reconcile(request)

		//then
		require.NoError(t, err)
		assertGitSourceAnalysis(t, client, "", langs, buildType(build.Ruby, "Gemfile"))
	}
}

func TestReconcileGitSourceAnalysisFromLocalRepoWithSshKey(t *testing.T) {
	//given
	defer gock.OffAll()
	allowedPubKey := test.PublicWithoutPassphrase(t, pathToTestDir)
	reset := test.RunKeySshServer(t, allowedPubKey)
	defer reset()

	dummyRepo := test.NewDummyGitRepo(t, repository.Master)
	dummyRepo.Commit("pom.xml", "mvnw", "src/main/java/Any.java", "pkg/main.go")

	for _, secretType := range []corev1.SecretType{corev1.SecretTypeOpaque, corev1.SecretTypeSSHAuth} {
		secret := test.NewSecret(secretType, map[string][]byte{
			"ssh-privatekey": test.PrivateWithoutPassphrase(t, pathToTestDir)})

		gs := test.NewGitSource(test.WithURL("ssh://git@localhost:2222" + dummyRepo.Path))
		gs.Spec.SecretRef = &v1alpha1.SecretRef{Name: test.SecretName}
		gsa := test.NewGitSourceAnalysis(test.GitSourceName)
		reconciler, request, client := PrepareClient(test.GitSourceAnalysisName,
			test.RegisterGvkObject(v1alpha1.SchemeGroupVersion, gs, gsa),
			test.RegisterGvkObject(corev1.SchemeGroupVersion, secret))

		//when
		_, err := reconciler.Reconcile(request)

		//then
		require.NoError(t, err)
		assertGitSourceAnalysis(t, client, "", test.S())
	}
}

func TestReconcileGitSourceAnalysisFromLocalRepoWithSshKeyWithPassphrase(t *testing.T) {
	//given
	defer gock.OffAll()
	allowedPubKey := test.PublicWithPassphrase(t, pathToTestDir)
	reset := test.RunKeySshServer(t, allowedPubKey)
	defer reset()

	dummyRepo := test.NewDummyGitRepo(t, repository.Master)
	dummyRepo.Commit("pom.xml", "mvnw", "src/main/java/Any.java", "pkg/main.go")

	for _, secretType := range []corev1.SecretType{corev1.SecretTypeOpaque, corev1.SecretTypeSSHAuth} {
		secret := test.NewSecret(secretType, map[string][]byte{
			"ssh-privatekey": test.PrivateWithPassphrase(t, pathToTestDir),
			"passphrase":     []byte("secret")})

		gs := test.NewGitSource(test.WithURL("ssh://git@localhost:2222" + dummyRepo.Path))
		gs.Spec.SecretRef = &v1alpha1.SecretRef{Name: test.SecretName}
		gsa := test.NewGitSourceAnalysis(test.GitSourceName)
		reconciler, request, client := PrepareClient(test.GitSourceAnalysisName,
			test.RegisterGvkObject(v1alpha1.SchemeGroupVersion, gs, gsa),
			test.RegisterGvkObject(corev1.SchemeGroupVersion, secret))

		//when
		_, err := reconciler.Reconcile(request)

		//then
		require.NoError(t, err)
		assertGitSourceAnalysis(t, client, "", test.S())
	}
}

func assertGitSourceAnalysis(t *testing.T, client client.Client, errorMsg string, langs test.SliceOfStrings, buildTypes ...typeWithFiles) {
	gitSourceAnalysis := &v1alpha1.GitSourceAnalysis{}
	err := client.Get(context.TODO(), newNamespacedName(test.Namespace, test.GitSourceAnalysisName), gitSourceAnalysis)
	require.NoError(t, err)

	if errorMsg != "" {
		assert.Contains(t, gitSourceAnalysis.Status.Error, errorMsg)
	} else {
		assert.Empty(t, gitSourceAnalysis.Status.Error)
	}

	buildEnvStats := gitSourceAnalysis.Status.BuildEnvStatistics
	require.Len(t, buildEnvStats.DetectedBuildTypes, len(buildTypes))
	for _, bt := range buildTypes {
		tool, files := bt()
		test.AssertContainsBuildTool(t, buildEnvStats.DetectedBuildTypes, tool.Name, tool.Language, files...)
	}

	if langs != nil {
		require.Len(t, buildEnvStats.SortedLanguages, len(langs()))
		for _, lang := range langs() {
			assert.Contains(t, buildEnvStats.SortedLanguages, lang)
		}
	} else {
		assert.Empty(t, buildEnvStats.SortedLanguages)
	}
}

type typeWithFiles func() (build.Tool, []string)

func buildType(tool build.Tool, files ...string) typeWithFiles {
	return func() (build.Tool, []string) {
		return tool, files
	}
}

func matchBasicAuth(usernameWithPassword string) test.GockModifier {
	return func(mock *gock.Request) {
		mock.MatchHeader("Authorization",
			fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte(usernameWithPassword))))
	}
}

func matchToken(token string) test.GockModifier {
	return func(mock *gock.Request) {
		mock.MatchHeader("Authorization", "Bearer "+token)
	}
}

func PrepareClient(name string, gvkObjects ...test.GvkObject) (*ReconcileGitSourceAnalysis, reconcile.Request, client.Client) {
	// Create a fake client to mock API calls.
	cl, s := test.PrepareClient(gvkObjects...)

	// Create a ReconcileToolChainEnabler object with the scheme and fake client.
	r := &ReconcileGitSourceAnalysis{client: cl, scheme: s}
	req := test.NewReconcileRequest(name)

	return r, req, cl
}
