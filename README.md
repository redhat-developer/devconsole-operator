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
$ operator-sdk add api --api-version=devconsole.openshift.io/v1alpha1 --kind=AppService
```

Add a new controller that watches for AppService

```sh
$ operator-sdk add controller --api-version=devconsole.openshift.io/v1alpha1 --kind=AppService
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