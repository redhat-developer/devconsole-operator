package operatorsource

import (
	"fmt"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	"os"
	"testing"
	"time"
)

var (
	Client               = NewTestClient()
	namespace            = "openshift-operators"
	subName              = "devconsole"
	label                = "name=devconsole-operator"
	subscription, suberr = Client.GetSubscription(subName, namespace)
)

func Test_OperatorSource(t *testing.T) {
	pods, err := Client.GetPodByLabel(label, namespace)
	if err != nil {
		t.Fatal(err)
	}
	defer CleanUp(t, &pods.Items[0])
	retryInterval := time.Second * 10
	timeout := time.Second * 120

	err = Client.WaitForOperatorDeployment(t, pods.Items[0].Name, namespace, retryInterval, timeout)
	if err != nil {
		t.Fatal(err)
	} else {
		t.Run("subscription", func(t *testing.T) { Subscription(t) })
		t.Run("install plan", func(t *testing.T) { InstallPlan(t) })
		t.Run("operator pod", func(t *testing.T) { OperatorPod(t) })
	}
}

func Subscription(t *testing.T) {
	// 1) Verify that the subscription was created
	if suberr != nil {
		t.Fatal(suberr)
	}
	require.Equal(t, subName, subscription.Name)
	require.Equal(t, "installed-custom-openshift-operators", subscription.Spec.CatalogSource)
}

func InstallPlan(t *testing.T) {
	// 2) Find the name of the install plan
	installPlanName := subscription.Status.Install.Name
	installPlan, err := Client.GetInstallPlan(installPlanName, namespace)
	if err != nil {
		t.Fatal(err)
	}
	require.Equal(t, "devconsole-operator.v0.1.0", installPlan.Spec.ClusterServiceVersionNames[0])
	require.Equal(t, "Automatic", string(installPlan.Spec.Approval))
	if !installPlan.Spec.Approved {
		require.FailNow(t, "Install plan approved is false")
	}
}

func OperatorPod(t *testing.T) {
	// 3) Check operator pod status, fail status != Running
	pods, err := Client.GetPodByLabel(label, namespace)
	if err != nil {
		t.Fatal(err)
	}
	pod := pods.Items[0]
	require.Equal(t, pod.Status.Phase, corev1.PodRunning)
}

func CleanUp(t *testing.T, pod *corev1.Pod) {
	// Clean up resources
	operatorVersion := os.Getenv("DEVCONSOLE_OPERATOR_VERSION")

	err := Client.Delete("installplan", subscription.Status.Install.Name, namespace)
	if err != nil {
		t.Logf("Error: %v\n", err)
	}

	err = Client.Delete("catsrc", subscription.Spec.CatalogSource, namespace)
	if err != nil {
		t.Logf("Error: %v\n", err)
	}

	err = Client.Delete("sub", subName, namespace)
	if err != nil {
		t.Logf("Error: %v\n", err)
	}

	csv := fmt.Sprintf("devconsole-operator.v%s", operatorVersion)
	err = Client.Delete("csv", csv, namespace)
	if err != nil {
		t.Logf("Error: %v\n", err)
	}

	err = Client.Delete("pod", pod.Name, namespace)
	if err != nil {
		t.Logf("Error: %v\n", err)
	}
}
