package gitsource

import (
	"context"
	"github.com/redhat-developer/devconsole-api/pkg/apis/devconsole/v1alpha1"
	"github.com/redhat-developer/devconsole-git/pkg/git/repository"
	gslog "github.com/redhat-developer/devconsole-git/pkg/log"
	gittransport "gopkg.in/src-d/go-git.v4/plumbing/transport"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var (
	log         = logf.Log.WithName("controller_gitsource")
	unableReach = NewConnection("Unable to reach the URL", v1alpha1.Failed)
)

// Add creates a new GitSource Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileGitSource{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("gitsource-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource GitSource
	err = c.Watch(&source.Kind{Type: &v1alpha1.GitSource{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileGitSource{}

// ReconcileGitSource reconciles a GitSource object
type ReconcileGitSource struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a GitSource object and makes changes based on the state read
// and what is in the GitSource.Spec
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileGitSource) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling GitSource")

	// Fetch the GitSource instance
	gitSource := &v1alpha1.GitSource{}
	err := r.client.Get(context.TODO(), request.NamespacedName, gitSource)
	if err != nil {
		reqLogger.Error(err, "Error getting GitSource object")
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}
	gitSourceLogger := gslog.LogWithGSValues(reqLogger, gitSource)

	shouldUpdate := updateStatus(gitSourceLogger, gitSource)

	if shouldUpdate {
		err = r.client.Update(context.TODO(), gitSource)
		if err != nil {
			gitSourceLogger.Error(err, "Error updating GitSource object")
			if errors.IsNotFound(err) {
				// Request object not found, could have been deleted after reconcile request.
				// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
				// Return and don't requeue
				return reconcile.Result{}, nil
			}
			// Error updating the object - requeue the request.
			return reconcile.Result{}, err
		}
	}
	return reconcile.Result{}, nil
}

func updateStatus(log *gslog.GitSourceLogger, gitSource *v1alpha1.GitSource) (isUpdated bool) {
	if gitSource.Status.Connection.State != "" {
		return false
	}
	if gitSource.Status.State == "" {
		gitSource.Status.State = v1alpha1.Initializing
	}

	endpoint, err := gittransport.NewEndpoint(gitSource.Spec.URL)
	if err != nil {
		gitSource.Status.Connection = NewConnection("unable to parse the URL: "+err.Error(), v1alpha1.Failed)
	} else {
		if gitSource.Spec.SecretRef == nil {
			ok, err := repository.IsReachableWithBranch(log, gitSource.Spec.Ref, endpoint)

			if ok {
				gitSource.Status.Connection = NewConnection("", v1alpha1.OK)
			} else {
				gitSource.Status.Connection = NewConnection(err.Error(), v1alpha1.Failed)
			}
		}
	}
	return true
}

func NewConnection(errorMsg string, state v1alpha1.ConnectionState) v1alpha1.Connection {
	return v1alpha1.Connection{
		Error: errorMsg,
		State: state,
	}
}
