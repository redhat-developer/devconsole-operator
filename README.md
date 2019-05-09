# DevConsole Operator

[![Go Report Card](https://goreportcard.com/badge/github.com/redhat-developer/devconsole-operator)](https://goreportcard.com/report/github.com/redhat-developer/devconsole-operator)
[![Docker Repository on Quay](https://quay.io/repository/redhat-developer/devconsole-operator/status "Docker Repository on Quay")](https://quay.io/repository/redhat-developer/devconsole-operator)


DevConsole operator enables a developer-focused view in OpenShift 4.
It provides a view switcher to transition between the traditional
Kubernetes cluster administration console referred to as
Administrator, to this new Developer perspective.

This new Developer perspective provides a high-level abstraction over
Kubernetes and OpenShift primitives to allow developers to focus on
their application development.

## Key Features

The Developer perspective is still under active development.  These
are the main features that are getting developed:

* Add - The place to create and build the application using one of this method:

	- Importing source code from Git
	- Deploying an existing image
	- Browse a catalog to deploy or connect application services
	- Deploy quick-starters or samples

* Topology - The landing page that shows application structure and
  health in an easy-to-use diagram
* Builds - Lists OpenShift BuildConfig resources for the selected
  project
* Pipelines - Lists Tekton Pipeline resources for the selected project

## Development

This repository was initially bootstrapped using the [Operator Framework SDK][operator-sdk].
This project requires [Go] version 1.11 or above.

Here is the complete list of pre-requisites:

- [Operator SDK][operator-sdk] version 0.7.0
- [dep][dep_tool] version 0.5.1
- [git][git_tool]
- [go][go_tool] version 1.11 or above
- [docker][docker_tool] version 17.03+
- [kubectl][kubectl_tool] version v1.11.0+ or [oc] version 3.11
- Access to OpenShift 4 cluster

### Build
```
make build
```
### Test
* run unit test:
```
make test
```
* run e2e test:
For running e2e tests, have minishift started.
```
make test-e2e-local
```
> Note: e2e test will deploy operator in project `devconsole-e2e-test`, if your tests timeout and you wan to debug:
> - oc project devconsole-e2e-test
> - oc get deployment,pod
> - oc logs pod/devconsole-operator-5b4bbc7d-4p7hr

## Deployment

### Set up Minishift (one-off)
* create a new profile to test the operator
```
minishift profile set devconsole
```
* enable the admin-user add-on
```
minishift addon enable admin-user
```
* start the instance
```
make minishift-start
```
> NOTE: this setup should be deprecated in favor of [OCP4 install]().

### Deploy the operator in dev mode

* In dev mode, simply run your operator locally:
```
make local
```
> NOTE: To watch all namespaces, `APP_NAMESPACE` is set to empty string. 
If a specific namespace is provided only that project will watched. 
As we reuse `openshift`'s imagestreams for build, we need to access all namespaces.

* Make sure minishift is running 
* Clean previously created resources
```
make deploy-clean
```
* Deploy CR
```
make deploy-test
```
* See the newly created resources
```
oc get all,dc,svc,dc,bc,route,cp,gitsource,gitsourceanalysis
```

### Deploy the operator with Deployment yaml

* (optional) minishift internal registry
Build the operator's controller image and make it available in internal registry
```
oc new-project devconsole
eval $(minishift docker-env)
operator-sdk build $(minishift openshift registry)/devconsole/devconsole-operator
```
> NOTE: In `operator.yaml` replace `imagePullPolicy: Always` with `imagePullPolicy: IfNotPresent` 
for local dev to avoid pulling image and be able to use docker cached image instead.
 
* deploy cr, role and rbac in `devconsole` namespace
```
make deploy-rbac
make deploy-crd
make deploy-operator
```
> NOTE: make sure `deploy/operator.yaml` points to your local image: `172.30.1.1:5000/devconsole/devconsole-operator:latest`

* watch the operator's pod
```
oc logs pod/devconsole-operator-5b4bbc7d-89crs -f
```

* in a different shell, test CR in different project (`local-test`)
```
make deploy-test
```
> Note: usee `make deploy-clean` to delete `local-test` project and start from fresh.

* check if the resources are created
```
oc get all,dc,svc,dc,bc,route,cp,gitsource,gitsourceanalysis
```
## Directory layout

Please consult [the documentation](https://github.com/operator-framework/operator-sdk/blob/master/doc/project_layout.md) in order to learn about this project's structure: 

|File/Folders  |Purpose |
|--------------|--------|
| cmd          | Contains `manager/main.go` which is the main program of the operator. This instantiates a new manager which registers all custom resource definitions under `pkg/apis/...` and starts all controllers under `pkg/controllers/...`.|
| pkg/apis | Contains the directory tree that defines the APIs of the Custom Resource Definitions(CRD). Users are expected to edit the `pkg/apis/<group>/<version>/<kind>_types.go` files to define the API for each resource type and import these packages in their controllers to watch for these resource types.|
| pkg/controller | This pkg contains the controller implementations. Users are expected to edit the `pkg/controller/<kind>/<kind>_controller.go` to define the controller's reconcile logic for handling a resource type of the specified `kind`.|
| build | Contains the `Dockerfile` and build scripts used to build the operator.|
| deploy | Contains various YAML manifests for registering CRDs, setting up [RBAC](https://kubernetes.io/docs/reference/access-authn-authz/rbac/), and deploying the operator as a Deployment.|
| Gopkg.toml Gopkg.lock | The [dep](https://github.com/golang/dep) manifests that describe the external dependencies of this operator.|
| vendor | The golang [Vendor](https://golang.org/cmd/go/#hdr-Vendor_Directories) folder that contains the local copies of the external dependencies that satisfy the imports of this project. [dep](https://github.com/golang/dep) manages the vendor directly.|


## Enabling the Developer  perspective in OpenShift

The frontend can check for the presence of the devconsole CRDs using the Kubernetes API.  Check for [the existence of a Custom Resource Definitions](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.13/#list-customresourcedefinition-v1beta1-apiextensions) with name as `gitsources.devconsole.openshift.io`.  If it exists, it will enable the Developer perspective in the Openshift Console.

Refer to OLM test [README](test/README.md) to install the DevOps Console operator.

[operator-sdk]: https://github.com/operator-framework/operator-sdk
[dep_tool]: https://golang.github.io/dep/docs/installation.html
[git_tool]: https://git-scm.com/downloads
[kubectl_tool]: https://kubernetes.io/docs/tasks/tools/install-kubectl/
[oc]: https://www.okd.io/download.html
[go_tool]: https://golang.org/dl/
[docker_tool]: https://docs.docker.com/install/
[Go]: https://golang.org
