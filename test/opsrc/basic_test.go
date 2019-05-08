package opsrc

import (
    "bytes"
    "fmt"
    "github.com/stretchr/testify/assert"   
    "log"
    "os"
    "os/exec"
    "strings"
    "testing"
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
    err, out, errout := Shellout("oc login -u " + os.Getenv("OC_LOGIN_USERNAME") + " -p " + os.Getenv("OC_LOGIN_PASSWORD"))
    if err != nil {
        log.Printf("error: %v\n", err)
        fmt.Println("stderr: " + errout)
    } else {
        // fmt.Println("stdout: " + out)
    }

    // 1) Verify that the subscription was created
    err, out, errout = Shellout("oc get sub devconsole -n openshift-operators")
    if err != nil {
        log.Printf("error: %v\n", err)
    } else {
        // fmt.Println("stdout: " + out)
        // fmt.Println("stderr: " + errout)
        assert.True(t, strings.Contains(out, "devconsole"), "Expecting the subscription name to be found")
        assert.True(t, strings.Contains(out, "installed-custom-openshift-operators"), "Expecting the subscription namespace to be found")
    }

    // 2) Find the name of the install plan
    err, out, errout = Shellout("oc get sub devconsole -n openshift-operators -o jsonpath='{.status.installplan.name}'")
    if err != nil {
        log.Printf("error: %v\n", err)
    } else {
        // fmt.Println("stdout: " + out)
        os.Setenv("INSTALL_PLAN", out)
        // fmt.Println("stderr: " + errout)
        // fmt.Println("INSTALL_PLAN = ", os.Getenv("INSTALL_PLAN"))
    }

    // 3) Verify the install plan
    err, out, errout = Shellout("oc get installplan " + out + " -n openshift-operators")
    if err != nil {
        log.Printf("error: %v\n", err)
    } else {
        // fmt.Println("stdout: " + out)
        // fmt.Println("stderr: " + errout) 
        assert.True(t, strings.Contains(out, os.Getenv("INSTALL_PLAN")), "Expecting the Install Plan name to be found")
        assert.True(t, strings.Contains(out, "devconsole-operator.v0.1.0"), "Expecting the Operator release to be found")
        assert.True(t, strings.Contains(out, "Automatic"), "Expecting the approval method to be found")
        assert.True(t, strings.Contains(out, "true"), "Expecting the approved state to be found")
    }

    // Verify that the operator's pod is running
    err, out, errout = Shellout("oc get pods  -l name=devconsole-operator -n openshift-operators -o jsonpath='{.items[*].status.phase}'")
    if err != nil {
        log.Printf("error: %v\n", err)
    } else {
        // fmt.Println("stdout: " + out)
        // fmt.Println("stderr: " + errout)
        assert.True(t, strings.Contains(out, "Running"), "Expecting the state of the Operator pod to be running")
    }

}

