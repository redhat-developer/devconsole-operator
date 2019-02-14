## Quick Start

Install the operator-sdk CLI:

```sh
$ mkdir -p $GOPATH/src/github.com/operator-framework
$ cd $GOPATH/src/github.com/operator-framework
$ git clone https://github.com/operator-framework/operator-sdk
$ cd operator-sdk
$ git checkout master
$ make dep
$ make install
```

Add a new API for the custom resource AppService

```sh
$ operator-sdk add api --api-version=devopsconsole-operator.openshift.io/v1alpha1 --kind=AppService
```

Add a new controller that watches for AppService

```sh
$ operator-sdk add controller --api-version=devopsconsole-operator.openshift.io/v1alpha1 --kind=AppService
```

Apply the app-operator CRD.

```sh
$ kubectl apply -f deploy/crds/app_v1alpha1_appservice_crd.yaml
```

Set the OPERATOR_NAME variable

```sh
$ export OPERATOR_NAME=app-operator
```

Run the operator from outside the Minishift environment.

```sh
$ operator-sdk up local --namespace myproject
```

Your operator is now watching for the existence of an object that matches: GroupVersionKind(app.example.com/v1alpha1, Kind=App)

Apply the provided CR.yaml to your cluster. This should trigger the default logic specified in the handler.

```sh
$ kubectl create -f deploy/examples/app_v1alpha1_appservice_cr.yaml
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
