package gitsourceanalysis

import (
	"context"

	"github.com/go-logr/logr"

	"github.com/redhat-developer/devconsole-api/pkg/apis/devconsole/v1alpha1"

	"github.com/redhat-developer/devconsole-git/pkg/git"
	"github.com/redhat-developer/devconsole-git/pkg/git/detector"
	"github.com/redhat-developer/devconsole-git/pkg/log"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var controllerLogger = logf.Log.WithName("controller_gitsourceanalysis")

// Add creates a new GitSourceAnalysis Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileGitSourceAnalysis{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("gitsourceanalysis-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource GitSourceAnalysis
	err = c.Watch(&source.Kind{Type: &v1alpha1.GitSourceAnalysis{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileGitSourceAnalysis{}

// ReconcileGitSourceAnalysis reconciles a GitSourceAnalysis object
type ReconcileGitSourceAnalysis struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a GitSourceAnalysis object and makes changes based on the state read
// and what is in the GitSourceAnalysis.Spec
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileGitSourceAnalysis) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := controllerLogger.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling GitSourceAnalysis")

	// Fetch the GitSourceAnalysis instance
	gsAnalysis := &v1alpha1.GitSourceAnalysis{}
	err := r.client.Get(context.TODO(), request.NamespacedName, gsAnalysis)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		reqLogger.Error(err, "There was an error while reading the GitSourceAnalysis object")
		return reconcile.Result{}, err
	}

	if gsAnalysis.Status.Analyzed {
		reqLogger.WithValues("git-source", gsAnalysis.Spec.GitSourceRef).
			Info("Skipping GitSourceAnalysis as it was already analyzed")
		return reconcile.Result{}, err
	}

	return analyze(reqLogger, r.client, gsAnalysis, request.Namespace)
}

func analyze(logger logr.Logger, client client.Client, gsAnalysis *v1alpha1.GitSourceAnalysis, namespace string) (reconcile.Result, error) {
	gitSource := &v1alpha1.GitSource{}
	err := client.Get(context.TODO(), newNsdName(namespace, gsAnalysis.Spec.GitSourceRef.Name), gitSource)
	if err != nil {
		gsAnalysis.Status.Error = "Failed to fetch the input source"
		logger.WithValues("git-source", gsAnalysis.Spec.GitSourceRef).
			Error(err, "There was an error while reading the GitSource object")

	} else {
		buildEnvStats, err := analyzeGitSource(log.LogWithGSValues(logger, gitSource), client, gitSource, namespace)
		if err != nil {
			gsAnalysis.Status.Error = err.Error()
		} else {
			gsAnalysis.Status.BuildEnvStatistics = *buildEnvStats
		}
	}

	gsAnalysis.Status.Analyzed = true
	err = client.Update(context.TODO(), gsAnalysis)
	if err != nil {
		log.LogWithGSValues(logger, gitSource).Error(err, "Error updating GitSourceAnalysis object")
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error updating the object - requeue the request.
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

func analyzeGitSource(logger *log.GitSourceLogger, client client.Client, gitSource *v1alpha1.GitSource, namespace string) (*v1alpha1.BuildEnvStats, error) {
	logger.Info("Analyzing GitSource")

	// Fetch the GitSource secret
	gitSecretProvider, err := git.NewGitSecretProvider(client, namespace, gitSource)
	if err != nil {
		logger.WithValues("secret", gitSource.Spec.SecretRef.Name).
			Error(err, "Error reading the secret object")
		return nil, err

	} else {
		buildEnvStats, err := detector.DetectBuildEnvironments(logger, gitSource, gitSecretProvider)
		if err != nil {
			logger.Error(err, "Error detecting build types")
		}
		return buildEnvStats, err
	}
}

func newNsdName(namespace, name string) types.NamespacedName {
	return types.NamespacedName{Namespace: namespace, Name: name}
}
