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

func Shellout(command string) (error, string, string) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd := exec.Command(ShellToUse, "-c", command)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	return err, stdout.String(), stderr.String()
}

func Test_OperatorSource_oc_commands(t *testing.T) {

	// Start - Login to oc
	err, out, _ := Shellout("oc login -u " + os.Getenv("OC_LOGIN_USERNAME") + " -p " + os.Getenv("OC_LOGIN_PASSWORD"))
	if err != nil {
		t.Fatalf("error: %v\n", err)
	}

	// 1) Verify that the subscription was created
	err, out, _ = Shellout("oc get sub devconsole -n openshift-operators")
	if err != nil {
		t.Fatalf("error: %v\n", err)
	} else {
		// t.Logf("stdout: %s\n", out)
		// t.Logf("stderr: %s\n", errout)
		require.True(t, strings.Contains(out, "devconsole"), "Expecting the subscription name to be found")
		require.True(t, strings.Contains(out, "installed-custom-openshift-operators"), "Expecting the subscription namespace to be found")
	}

	// 2) Find the name of the install plan
	err, out, _ = Shellout("oc get sub devconsole -n openshift-operators -o jsonpath='{.status.installplan.name}'")
	var installPlan string
	if err != nil {
		t.Fatalf("error: %v\n", err)
	} else {
		// t.Logf("stdout: %s\n", out)
		// t.Logf("stderr: %s\n", errout)
		installPlan = out
	}

	// 3) Verify the install plan
	err, out, _ = Shellout(fmt.Sprintf("oc get installplan %s -n openshift-operators", installPlan))
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

	// Verify that the operator's pod is running
	err, out, _ = Shellout("oc get pods  -l name=devconsole-operator -n openshift-operators -o jsonpath='{.items[*].status.phase}'")
	if err != nil {
		t.Fatalf("error: %v\n", err)
	} else {
		// t.Logf("stdout: %s\n", out)
		// t.Logf("stderr: %s\n", errout)
		require.True(t, strings.Contains(out, "Running"), "Expecting the state of the Operator pod to be running")
	}

}
