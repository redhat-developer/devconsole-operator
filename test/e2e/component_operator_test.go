package e2e

import (
	"context"
	"fmt"
	framework "github.com/operator-framework/operator-sdk/pkg/test"
	"github.com/operator-framework/operator-sdk/pkg/test/e2eutil"
	"github.com/redhat-developer/devopsconsole-operator/pkg/apis"
	componentsv1alpha1 "github.com/redhat-developer/devopsconsole-operator/pkg/apis/devopsconsole/v1alpha1"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	corev1client "k8s.io/client-go/kubernetes/typed/core/v1"
	"os"
	"testing"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

const (
	retryInterval        = time.Second * 5
	timeout              = time.Second * 60
	cleanupRetryInterval = time.Second * 1
	cleanupTimeout       = time.Second * 5
)

type ComponentTestSuite struct {
	suite.Suite
	namespace string
	framework *framework.Framework
	client    *corev1client.CoreV1Client
	ctx       *framework.TestCtx
}

func (suite *ComponentTestSuite) SetupSuite() {
	suite.framework = framework.Global

	coreclient, err := corev1client.NewForConfig(framework.Global.KubeConfig)
	if err != nil {
		panic("failed to create new client")
	}
	suite.client = coreclient

	suite.ctx = framework.NewTestCtx(suite.T())

	namespace, err := suite.ctx.GetNamespace()
	require.NoError(suite.T(), err, "failed to get namespace where operator needs to run")
	suite.namespace = namespace
	suite.T().Log(fmt.Sprintf("namespace: %s", suite.namespace))

	// Register types with framework scheme
	componentList := &componentsv1alpha1.ComponentList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Component",
			APIVersion: "devopsconsole.openshift.io/v1alpha1",
		},
	}
	err = framework.AddToFrameworkScheme(apis.AddToScheme, componentList)
	if err != nil {
		suite.T().Fatalf("failed to add custom resource scheme to framework: %v", err)
	}

	err = suite.ctx.InitializeClusterResources(&framework.CleanupOptions{TestContext: suite.ctx, Timeout: cleanupTimeout, RetryInterval: cleanupRetryInterval})
	require.NoError(suite.T(), err, "failed to initialize cluster resources")
	suite.T().Log("Initialized cluster resources")

	// wait for component-operator to be ready
	err = e2eutil.WaitForDeployment(suite.T(), suite.framework.KubeClient, suite.namespace, "devopsconsole-operator", 1, retryInterval, timeout)
	require.NoError(suite.T(), err, "failed while waiting for operator deployment")

	suite.T().Log("component-operator is ready and running state")
}

// ComponentTest does e2e test as per operator-sdk documentation
// https://github.com/operator-framework/operator-sdk/blob/cc7b175/doc/test-framework/writing-e2e-tests.md
func (suite *ComponentTestSuite) TestComponent() {

	suite.T().Log(fmt.Sprintf("namespace: %s", suite.namespace))
	// create a Component custom resource
	cr := &componentsv1alpha1.Component{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Component",
			APIVersion: "devopsconsole.openshift.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "mycomp",
			Namespace: suite.namespace,
		},
		Spec: componentsv1alpha1.ComponentSpec{
			BuildType: "nodejs",
			Codebase:  "https://github.com/nodeshift-starters/nodejs-rest-http-crud",
		},
	}
	// use TestCtx's create helper to create the object and add a cleanup function for the new object
	err := suite.framework.Client.Create(context.TODO(), cr, &framework.CleanupOptions{TestContext: suite.ctx, Timeout: cleanupTimeout, RetryInterval: cleanupRetryInterval})
	require.NoError(suite.T(), err, "failed to create custom resource of kind `Component`")

	suite.T().Run("retrieve component and verify related resources are created", func(t *testing.T) {
		err = suite.framework.Client.Get(context.TODO(), types.NamespacedName{Name: "mycomp", Namespace: suite.namespace}, cr)
		require.NoError(t, err, "failed to retrieve custom resource of kind `Component`")
		require.Equal(t, "https://github.com/nodeshift-starters/nodejs-rest-http-crud", cr.Spec.Codebase)
		require.Equal(t, "nodejs", cr.Spec.BuildType)
	})
}

func (suite *ComponentTestSuite) TearDownSuite() {
	err := suite.client.Namespaces().Delete(suite.namespace, &metav1.DeleteOptions{})
	if err != nil {
		panic("failed to delete test namespace")
	}

	os.Unsetenv("TEST_NAMESPACE")
	suite.ctx.Cleanup()

	suite.T().Log("teardown complete")
}

func TestComponentTestSuite(t *testing.T) {
	suite.Run(t, new(ComponentTestSuite))
}
