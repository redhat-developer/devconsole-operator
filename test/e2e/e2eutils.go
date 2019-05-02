package e2e

import (
	goctx "context"
	"time"

	"github.com/operator-framework/operator-sdk/pkg/test"
	devconsoleapi "github.com/redhat-developer/devconsole-api/pkg/apis/devconsole/v1alpha1"

	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
)

// WaitUntilGitSourceReconcile waits execution until controller finishes reconciling.
func WaitUntilGitSourceReconcile(t *test.Framework, nsd types.NamespacedName) error {
	var err error
	err = wait.Poll(time.Second*5, time.Minute*1, func() (bool, error) {
		var gitSource devconsoleapi.GitSource
		err = t.Client.Get(goctx.TODO(), nsd, &gitSource)
		if err != nil {
			return false, err
		}
		return (gitSource.Status.Connection.State != ""), nil
	})
	return err
}

// WaitUntilGitSourceAnalyzeReconcile waits execution until controller finishes reconciling.
func WaitUntilGitSourceAnalyzeReconcile(t *test.Framework, nsd types.NamespacedName) error {
	var err error
	err = wait.Poll(time.Second*5, time.Minute*1, func() (bool, error) {
		var gitSourceAnalysis devconsoleapi.GitSourceAnalysis
		err = t.Client.Get(goctx.TODO(), nsd, &gitSourceAnalysis)
		if err != nil {
			return false, err
		}
		return gitSourceAnalysis.Status.Analyzed, nil
	})
	return err
}
