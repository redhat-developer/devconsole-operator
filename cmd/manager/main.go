package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"runtime"

	appsv1 "github.com/openshift/api/apps/v1"
	buildv1 "github.com/openshift/api/build/v1"
	imagev1 "github.com/openshift/api/image/v1"
	routev1 "github.com/openshift/api/route/v1"
	"github.com/operator-framework/operator-sdk/pkg/k8sutil"
	"github.com/operator-framework/operator-sdk/pkg/leader"
	"github.com/operator-framework/operator-sdk/pkg/ready"
	sdkVersion "github.com/operator-framework/operator-sdk/version"
	"k8s.io/apimachinery/pkg/api/errors"

	"github.com/redhat-developer/devopsconsole-operator/pkg/apis"
	devopsconsolev1alpha1 "github.com/redhat-developer/devopsconsole-operator/pkg/apis/devopsconsole/v1alpha1"
	"github.com/redhat-developer/devopsconsole-operator/pkg/controller"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/runtime/signals"
)

var log = logf.Log.WithName("cmd")

func printVersion() {
	log.Info(fmt.Sprintf("Go Version: %s", runtime.Version()))
	log.Info(fmt.Sprintf("Go OS/Arch: %s/%s", runtime.GOOS, runtime.GOARCH))
	log.Info(fmt.Sprintf("operator-sdk Version: %v", sdkVersion.Version))
}

func main() {
	flag.Parse()

	// The logger instantiated here can be changed to any logger
	// implementing the logr.Logger interface. This logger will
	// be propagated through the whole operator, generating
	// uniform and structured logs.
	logf.SetLogger(logf.ZapLogger(false))

	printVersion()

	namespace, err := k8sutil.GetWatchNamespace()
	if err != nil {
		log.Error(err, "failed to get watch namespace")
		os.Exit(1)
	}
	// Get a config to talk to the apiserver
	cfg, err := config.GetConfig()
	if err != nil {
		log.Error(err, "")
		os.Exit(1)
	}

	// Become the leader before proceeding
	err = leader.Become(context.TODO(), "devopsconsole-operator-lock")
	if err != nil {
		log.Error(err, "")
		os.Exit(1)
	}

	r := ready.NewFileReady()
	err = r.Set()
	if err != nil {
		log.Error(err, "")
		os.Exit(1)
	}
	defer func() {
		err := r.Unset()
		if err != nil {
			log.Error(err, "")
			os.Exit(1)
		}
	}()

	// Create a new Cmd to provide shared dependencies and start components
	mgr, err := manager.New(cfg, manager.Options{Namespace: namespace})
	if err != nil {
		log.Error(err, "")
		os.Exit(1)
	}

	log.Info("Registering Components.")

	// Setup Scheme for all resources
	// Setup Scheme for all resources
	if err := apis.AddToScheme(mgr.GetScheme()); err != nil {
		log.Error(err, "")
		os.Exit(1)
	}
	if err := imagev1.AddToScheme(mgr.GetScheme()); err != nil {
		log.Error(err, "")
		os.Exit(1)
	}
	if err := buildv1.AddToScheme(mgr.GetScheme()); err != nil {
		log.Error(err, "")
		os.Exit(1)
	}
	if err := appsv1.AddToScheme(mgr.GetScheme()); err != nil {
		log.Error(err, "")
		os.Exit(1)
	}

	if err := routev1.AddToScheme(mgr.GetScheme()); err != nil {
		log.Error(err, "")
		os.Exit(1)
	}

	// Setup all Controllers
	if err := controller.AddToManager(mgr); err != nil {
		log.Error(err, "")
		os.Exit(1)
	}

	// Create a cluster role and binding for it.
	// This is necessary because we want all users to have read access to the
	// installer CR.
	// Currently, adding clusterroles to groups/users is not supported by CSV.
	// That's why the clusterRole and binding are created over here manually
	if err := createClusterRoleAndBinding(mgr.GetClient()); err != nil {
		log.Error(err, "")
		os.Exit(1)
	}
	// Create a new "installer". This will trigger the deployment and create route
	c := mgr.GetClient()
	installer := &devopsconsolev1alpha1.Installer{}
	installer.ObjectMeta.Name = "main-installer"
	installer.ObjectMeta.Namespace = namespace
	err = c.Create(context.TODO(), installer)
	if err != nil && !errors.IsAlreadyExists(err) {
		log.Error(err, "failed to create installer resource")
		os.Exit(1)
	}

	log.Info("Starting the Cmd.")

	// Start the Cmd
	if err := mgr.Start(signals.SetupSignalHandler()); err != nil {
		log.Error(err, "manager exited non-zero")
		os.Exit(1)
	}
}

// createClusterRoleAndBinding creates new clusterrole and it's binding.
// This cluster role allows all authenticated users to read the status of
// `installer` CR
func createClusterRoleAndBinding(c client.Client) error {
	clusterRole := getClusterRole()
	if err := c.Update(context.TODO(), clusterRole); err != nil {
		return err
	}
	clusterRoleBinding := getClusterRoleBinding()
	if err := c.Update(context.TODO(), clusterRoleBinding); err != nil {
		return err
	}
	return nil
}

func getClusterRoleBinding() *rbacv1.ClusterRoleBinding {
	return &rbacv1.ClusterRoleBinding{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "ClusterRoleBinding",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "installer-clusterrolebinding",
		},
		Subjects: []rbacv1.Subject{
			{
				Kind: "Group",
				Name: "system:authenticated",
			},
		},
		RoleRef: rbacv1.RoleRef{
			Name: "devconsole-installer-clusterrole",
			Kind: "ClusterRole",
		},
	}
}

func getClusterRole() *rbacv1.ClusterRole {
	return &rbacv1.ClusterRole{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "ClusterRole",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "devconsole-installer-clusterrole",
		},
		Rules: []rbacv1.PolicyRule{
			{
				Verbs:     []string{"get", "list"},
				APIGroups: []string{"devopsconsole.openshift.io"},
				Resources: []string{"installers"},
			},
		},
	}
}
