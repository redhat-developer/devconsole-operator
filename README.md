

# DevConsole Operator


[![Go Report Card](https://goreportcard.com/badge/github.com/redhat-developer/devconsole-operator)](https://goreportcard.com/report/github.com/redhat-developer/devconsole-operator)
[![Docker Repository on Quay](https://quay.io/repository/redhat-developer/devconsole-operator/status "Docker Repository on Quay")](https://quay.io/repository/redhat-developer/devconsole-operator)
[![Docker Repository on Quay](https://quay.io/repository/redhat-developer/operator-registry/status "Docker Repository on Quay")](https://quay.io/repository/redhat-developer/operator-registry)

## Overview
The DevConsole operator enables a developer-focused view in the OpenShift 4 web
console. It provides a view switcher to transition between Administrator, the
traditional administration focused console, to a new Developer perspective.

This new Developer perspective provides a high-level of abstraction over
Kubernetes and OpenShift primitives to allow developers to focus on their
application development.

## Key Features

The Developer perspective is under active development. These are the main
features being developed:

* **Add**: Use this page to create and build an application using one of the following
methods:

    - Import source code from Git
    - Deploy an existing image
    - Browse the catalog to deploy or connect application services
    - Deploy quick-starts or samples

* **Topology**: The landing page that shows the application structure and health in an
 easy-to-use graphic representation.
* **Builds**: This page lists the OpenShift BuildConfig resources for the selected project.
* **Pipelines**: This page lists the Tekton Pipeline resources for the selected project.

## Installing the latest console with the developer perspective as a non-admin user

To install the latest console with the developer perspective:

1. Clone the [devconsole
repository](https://github.com/redhat-developer/devconsole-operator) locally.
1. Change directory to the `hack/install_devconsole` directory and run the script:
    ```
    sh consoledeveloper.sh
    ```

    The script:
      * Installs the latest console with the developer perspective
      * Installs the devconsole operator or prompts you if it already exists
      * Creates a non-admin user with the suitable rolebinding
      * Prompts you with the credentials to log in to the console

1. Log in and create a new project.
1. Run `oc get csvs` in the suitable namespace to see the installed operator.

## Development

This repository was initially bootstrapped using the [Operator Framework SDK][operator-sdk] and the project requires [Go] version 1.11 or above.

**Prerequisites**:

- [Operator SDK][operator-sdk] version 0.7.0
- [dep][dep_tool] version 0.5.1
- [git][git_tool]
- [go][go_tool] version 1.11 or above
- [docker][docker_tool] version 17.03+
- [kubectl][kubectl_tool] version v1.11.0+ or [oc] version 3.11
- Access to OpenShift 4 cluster

### Build
To build the operator use:
```
make build
```
### Test
* To run unit test use:
    ```
    make test
    ```
* To run e2e test:

    Start Minishift and run:
    ```
    make test-e2e-local
    ```

    **Note:**
The e2e test deploys the operator in the project `devconsole-e2e-test`. If your
tests timeout and you want to debug, run:
    ```
    oc project devconsole-e2e-test
    oc get deployment,pod
    oc logs pod/devconsole-operator-<pod_number>
    ```


## Deployment

**Prerequisites:**

Set up Minishift (a one time task):

1. Create a new profile to test the operator.
    ```
    minishift profile set devconsole
    ```
1. Enable the admin-user add-on.
    ```
    minishift addon enable admin-user
    ```
1. Start the instance.
    ```
    make minishift-start
    ```
**NOTE:** Eventually this setup will be deprecated in favor of [Code Ready Containers]() installation.

### Deploying the operator in dev mode

1. In dev mode, simply run your operator locally:
    ```
    make local
    ```
    **Note:** To watch all namespaces, `APP_NAMESPACE` is set to empty string.
    If a specific namespace is provided only that project is watched.
    As we reuse `openshift`'s imagestreams for build, we need to access all namespaces.

1. Make sure minishift is running.
1. Clean previously created resources:
    ```
    make deploy-clean
    ```
1. Deploy the CR.
    ```
    make deploy-test
    ```
1. Check the freshly created resources.
    ```
    oc get all,dc,svc,dc,bc,route,cp,gitsource,gitsourceanalysis
    ```

### Deploying the operator with Deployment yaml

1. (Optional) Build the operator image and make it available in the Minishift internal registry.
    ```
    oc new-project devconsole
     $(minishift docker-env)
    operator-sdk build $(minishift openshift registry)/devconsole/devconsole-operator
    ```
    **Note:** To avoid pulling the image and use the docker cached image instead for local dev, in the `operator.yaml`, replace `imagePullPolicy: Always` with `imagePullPolicy: IfNotPresent`.

1. Deploy the CR, role, and rbac in the `devconsole` namespace:
    ```
    make deploy-rbac
    make deploy-crd
    make deploy-operator
    ```
    **Note:** Make sure `deploy/operator.yaml` points to your local image: `172.30.1.1:5000/devconsole/devconsole-operator:latest`

1. Watch the operator pod:
    ```
    oc logs pod/devconsole-operator-5b4bbc7d-89crs -f
    ```

1. In a different shell, test CR in a different project (`local-test`):
    ```
    make deploy-test
    ```
    **Note:** Use `make deploy-clean` to delete `local-test` project and start fresh.

1. Check if the resources are created:
    ```
    oc get all,dc,svc,dc,bc,route,cp,gitsource,gitsourceanalysis
    ```

## Directory layout

See [Operator-SDK documentation](https://github.com/operator-framework/operator-sdk/blob/master/doc/project_layout.md) in order to learn about this project's structure:

|File/Folders  |Purpose |
|--------------|--------|
| cmd          | Contains `manager/main.go` which is the main program of the operator. This instantiates a new manager which registers all custom resource definitions under `pkg/apis/...` and starts all controllers under `pkg/controllers/...`.|
| pkg/apis | Contains the directory tree that defines the APIs of the Custom Resource Definitions(CRD). Users are expected to edit the `pkg/apis/<group>/<version>/<kind>_types.go` files to define the API for each resource type and import these packages in their controllers to watch for these resource types.|
| pkg/controller | Contains the controller implementations. Users are expected to edit the `pkg/controller/<kind>/<kind>_controller.go` to define the controller's reconcile logic for handling a resource type of the specified `kind`.|
| build | Contains the `Dockerfile` and build scripts used to build the operator.|
| deploy | Contains various YAML manifests for registering CRDs, setting up [RBAC](https://kubernetes.io/docs/reference/access-authn-authz/rbac/), and deploying the operator as a Deployment.|
| Gopkg.toml Gopkg.lock | The [dep](https://github.com/golang/dep) manifests that describe the external dependencies of this operator.|
| vendor | The golang [Vendor](https://golang.org/cmd/go/#hdr-Vendor_Directories) folder that contains the local copies of the external dependencies that satisfy the imports of this project. [dep](https://github.com/golang/dep) manages the vendor directly.|


## Enabling the Developer perspective in OpenShift

The frontend must check for [the presence of the devconsole Custom Resource
Definition](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.13/#list-customresourcedefinition-v1beta1-apiextensions)
named `gitsources.devconsole.openshift.io` using the Kubernetes API. This CRD
enables the Developer perspective in the OpenShift Console.

Refer to the OLM test [README](test/README.md) to run the end to end (E2E) tests.

[operator-sdk]: https://github.com/operator-framework/operator-sdk
[dep_tool]: https://golang.github.io/dep/docs/installation.html
[git_tool]: https://git-scm.com/downloads
[kubectl_tool]: https://kubernetes.io/docs/tasks/tools/install-kubectl/
[oc]: https://www.okd.io/download.html
[go_tool]: https://golang.org/dl/
[docker_tool]: https://docs.docker.com/install/
[Go]: https://golang.org
[Code Ready Containers]: https://github.com/code-ready/crc
