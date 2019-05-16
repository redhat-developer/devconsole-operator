package topology

import (
	"testing"

	"github.com/stretchr/testify/assert"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	coreFakeClient "k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/kubernetes/scheme"

	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func TestReconcile(t *testing.T) {
	t.Run("Check if Deployment and Services are getting created", func(t *testing.T) {
		objs := []runtime.Object{}
		cl := fake.NewFakeClient(objs...)
		s := scheme.Scheme
		coreFakeC := coreFakeClient.NewSimpleClientset(objs...)
		r := ReconcileService{
			client:     cl,
			coreClient: coreFakeC,
			scheme:     s,
		}
		_, err := r.Reconcile(reconcile.Request{
			NamespacedName: types.NamespacedName{
				Name:      ServiceName,
				Namespace: ServicesNamespace,
			},
		})
		if err != nil {
			t.Log(err)
			t.Fail()
		}
		assert.Nil(t, err, "Reconcile failed with error ")
		deployment, _ := r.coreClient.AppsV1().Deployments(ServicesNamespace).Get(ServiceName, metav1.GetOptions{})
		assert.NotNil(t, deployment, "Deployment should have created")
		assert.Equal(t, ServiceName, deployment.Name, "Deployment is not created with name "+ServiceName)
		assert.Equal(t, ServicesNamespace, deployment.Namespace, "Deployment is not created in expected namespace"+ServicesNamespace)
		service, _ := r.coreClient.CoreV1().Services(ServicesNamespace).Get(ServiceName, metav1.GetOptions{})
		assert.NotNil(t, service, "Service should have created")
		assert.Equal(t, ServiceName, deployment.Name, "Service is not created with name "+ServiceName)
		assert.Equal(t, ServicesNamespace, deployment.Namespace, "Service is not created in namespace "+ServicesNamespace)
	})
}
