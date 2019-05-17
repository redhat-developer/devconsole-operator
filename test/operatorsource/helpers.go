package operatorsource

import (
	"errors"
	"fmt"
	apis_v1alpha1 "github.com/operator-framework/operator-lifecycle-manager/pkg/api/apis/operators/v1alpha1"
	client "github.com/operator-framework/operator-lifecycle-manager/pkg/api/client"
	v1alpha1 "github.com/operator-framework/operator-lifecycle-manager/pkg/api/client/clientset/versioned/typed/operators/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"log"
	"os"
	"testing"
	"time"
)

//ClientSetK8sCoreAPI returns new Clientset for the given config, use to interact with K8s resources like pods
func ClientSetK8sCoreAPI(kubeconfig string) *kubernetes.Clientset {
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		log.Fatal(err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal(err)
	}
	return clientset
}

//ClientSet Creates clientset for given config, use to interact with custom resources
func ClientSet(kubeconfig string) v1alpha1.OperatorsV1alpha1Interface {
	client, err := client.NewClient(kubeconfig)
	if err != nil {
		log.Fatal(err)
	}
	return client.OperatorsV1alpha1()
}

// TestClient wraps all the clientsets required while testing
type TestClient struct {
	K8sClient      *kubernetes.Clientset
	OperatorClient v1alpha1.OperatorsV1alpha1Interface
}

//NewTestClient initialises the TestClient
func NewTestClient() *TestClient {
	kubeconfig := os.Getenv("KUBECONFIG")

	return &TestClient{
		K8sClient:      ClientSetK8sCoreAPI(kubeconfig),
		OperatorClient: ClientSet(kubeconfig),
	}
}

// GetPodByLabel is a function that takes label and namespace and returns the pod and error
func (tc *TestClient) GetPodByLabel(label string, namespace string) (*corev1.Pod, error) {

	pods, err := tc.K8sClient.CoreV1().Pods(namespace).List(metav1.ListOptions{LabelSelector: label})
	if err != nil {
		return nil, err
	}
	if len(pods.Items) == 0 {
		return nil, nil
	}
	return &pods.Items[0], nil
}

//GetSubscription returns subscription struct
func (tc *TestClient) GetSubscription(subName, namespace string) (*apis_v1alpha1.Subscription, error) {
	subscription, err := tc.OperatorClient.Subscriptions(namespace).Get(subName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return subscription, nil
}

//GetInstallPlan returns install plan struct
func (tc *TestClient) GetInstallPlan(installPlanName, namespace string) (*apis_v1alpha1.InstallPlan, error) {

	installPlan, err := tc.OperatorClient.InstallPlans(namespace).Get(installPlanName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return installPlan, nil
}

//Delete takes kind, name and namespace of the resource and deletes it. Returns an error if one occurs.
func (tc *TestClient) Delete(resource, name, namespace string) error {
	switch resource {
	case "subscription", "sub":
		err := tc.OperatorClient.Subscriptions(namespace).Delete(name, &metav1.DeleteOptions{})
		if err != nil {
			return err
		}
		return nil
	case "installplan":
		err := tc.OperatorClient.InstallPlans(namespace).Delete(name, &metav1.DeleteOptions{})
		if err != nil {
			return err
		}
		return nil
	case "catalogsource", "catsrc":
		err := tc.OperatorClient.CatalogSources(namespace).Delete(name, &metav1.DeleteOptions{})
		if err != nil {
			return err
		}
		return nil
	case "clusterserviceversion", "csv":
		err := tc.OperatorClient.ClusterServiceVersions(namespace).Delete(name, &metav1.DeleteOptions{})
		if err != nil {
			return err
		}
		return nil
	case "pod":
		err := tc.K8sClient.CoreV1().Pods(namespace).Delete(name, &metav1.DeleteOptions{})
		if err != nil {
			return err
		}
		return nil
	default:
		option := fmt.Sprintf("Invalid resource: %s", resource)
		return errors.New(option)

	}
}

//WaitForOperatorDeployment takes pod struct and wait till pods gets in runnig state
func (tc *TestClient) WaitForOperatorDeployment(t *testing.T, name, namespace string, retryInterval, timeout time.Duration) error {
	err := wait.Poll(retryInterval, timeout, func() (bool, error) {
		pod, err := tc.K8sClient.CoreV1().Pods(namespace).Get(name, metav1.GetOptions{})
		if err != nil {
			return false, err
		}
		if pod == nil {
			return false, errors.New("pod returned empty")
		}
		if pod.Status.Phase == corev1.PodRunning {
			return true, nil
		}
		t.Logf("Pod %s Status: %s", pod.Name, pod.Status.Phase)
		return false, nil
	})
	if err != nil {
		return err
	}
	return nil
}
