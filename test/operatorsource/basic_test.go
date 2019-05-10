package operatorsource

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

const ShellToUse = "bash"

func Shellout(command string) (string, string, error) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd := exec.Command(ShellToUse, "-c", command)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	return stdout.String(), stderr.String(), err
}

func Test_OperatorSource_oc_commands(t *testing.T) {

	defer CleanUp(t)

	t.Run("login", func(t *testing.T) { Login(t) })
	t.Run("subscription", func(t *testing.T) { Subscription(t) })
	t.Run("install plan", func(t *testing.T) { InstallPlan(t) })
	t.Run("operator pod", func(t *testing.T) { OperatorPod(t) })
}

func Login(t *testing.T) {
	// Start - Login to oc
	out, _, err := Shellout("oc login -u " + os.Getenv("OC_LOGIN_USERNAME") + " -p " + os.Getenv("OC_LOGIN_PASSWORD"))
	if err != nil {
		t.Fatalf("error: %v\n", err)
	} else {
		require.True(t, strings.Contains(out, "Login successful."), "Expecting successful login")
	}
}

func Subscription(t *testing.T) {
	// 1) Verify that the subscription was created
	out, _, err := Shellout("oc get sub devconsole -n openshift-operators")
	if err != nil {
		t.Fatalf("error: %v\n", err)
	} else {
		// t.Logf("stdout: %s\n", out)
		// t.Logf("stderr: %s\n", errout)
		require.True(t, strings.Contains(out, "devconsole"), "Expecting the subscription name to be found")
		require.True(t, strings.Contains(out, "installed-custom-openshift-operators"), "Expecting the subscription namespace to be found")
	}
}

func InstallPlan(t *testing.T) {
	// 2) Find the name of the install plan
	out, _, err := Shellout("oc get sub devconsole -n openshift-operators -o jsonpath='{.status.installplan.name}'")
	var installPlan string
	if err != nil {
		t.Fatalf("error: %v\n", err)
	} else {
		// t.Logf("stdout: %s\n", out)
		// t.Logf("stderr: %s\n", errout)
		installPlan = out
	}

	// 3) Verify the install plan
	out, _, err = Shellout(fmt.Sprintf("oc get installplan %s -n openshift-operators", installPlan))
	if err != nil {
		t.Fatalf("error: %v\n", err)
	} else {
		// t.Logf("stdout: %s\n", out)
		// t.Logf("stderr: %s\n", errout)
		require.True(t, strings.Contains(out, installPlan), "Expecting the Install Plan name to be found")
		require.True(t, strings.Contains(out, "devconsole-operator.v0.1.0"), "Expecting the Operator release to be found")
		require.True(t, strings.Contains(out, "Automatic"), "Expecting the approval method to be found")
		require.True(t, strings.Contains(out, "true"), "Expecting the approved state to be found")
	}
}

func OperatorPod(t *testing.T) {
	// Verify that the operator's pod is running
	out, _, err := Shellout("oc get pods  -l name=devconsole-operator -n openshift-operators -o jsonpath='{.items[*].status.phase}'")
	if err != nil {
		t.Fatalf("error: %v\n", err)
	} else {
		// t.Logf("stdout: %s\n", out)
		// t.Logf("stderr: %s\n", errout)
		require.True(t, strings.Contains(out, "Running"), "Expecting the state of the Operator pod to be running")
	}
}

func CleanUp(t *testing.T) {
	// Clean up resources
	operatorSourceName := os.Getenv("OPSRC_NAME")
	operatorVersion := os.Getenv("DEVCONSOLE_OPERATOR_VERSION")

	out, _, err := Shellout(fmt.Sprintf("oc delete opsrc %s -n openshift-marketplace", operatorSourceName))
	if err != nil {
		t.Logf("error: %v\n", err)
	} else {
		t.Logf(out)
	}

	out, _, err = Shellout("oc delete sub devconsole -n openshift-operators")
	if err != nil {
		t.Logf("error: %v\n", err)
	} else {
		t.Logf(out)
	}

	out, _, err = Shellout("oc delete catsrc installed-custom-openshift-operators -n openshift-operators")
	if err != nil {
		t.Logf("error: %v\n", err)
	} else {
		t.Logf(out)
	}

	out, _, err = Shellout("oc delete csc installed-custom-openshift-operators -n openshift-marketplace")
	if err != nil {
		t.Logf("error: %v\n", err)
	} else {
		t.Logf(out)
	}

	out, _, err = Shellout(fmt.Sprintf("oc delete csv devconsole-operator.v%s -n openshift-operators", operatorVersion))
	if err != nil {
		t.Logf("error: %v\n", err)
	} else {
		t.Logf(out)
	}
}
