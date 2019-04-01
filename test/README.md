## E2E Tests

### How to Run

#### TL;DR

```
make minishift-start
eval $(minishift docker-env)
make test-e2e
```

Make sure that you have minishift running. you can check it using `minishift status`. If not then start it using `make minishift-start` target.

After successfully starting minishift, configure your shell to use docker daemon from minishift using `eval $(minishift docker-env)`.

Now it's time to run E2E tests for `devconsole-operator` which will create it's required resources from `deploy/test/` on OpenShift use following command:

```
make test-e2e
```

This make target is building new docker image `$(DOCKER_REPO)/$(IMAGE_NAME):test`(e.g. `quay.io/openshiftio/devconsole-operator:test`) which is used in the operator's deployment manifests in e2e tests.

Also remember that it uses the `system:admin` account for creating all required resources from `deploy/test` directory.

## Steps to verify operator registry

### 1. Install OLM (not required for OpenShift 4)

If you are using OpenShift 3, install OLM with this command:

```
oc create -f https://raw.githubusercontent.com/operator-framework/operator-lifecycle-manager/master/deploy/upstream/quickstart/olm.yaml 
```

> NOTE: Alternately you can use `oc` command instead of `kubectl.`

### 2. Build and push the operator image to a public registry such as quay.io

Note: Instead of a public registry, the registry provided by OpenShift might work.

Checkout the `master` branch of [devconsole-operator](https://github.com/redhat-developer/devconsole-operator)

Then run these commands:

```
$ operator-sdk build quay.io/<username>/devconsole-operator
$ docker login -u <username> -p <password>  quay.io
$ docker push quay.io/<username>/devconsole-operator
```
> NOTE: make your repo public
When running the above command, substitute the `username` and `password` entries appropriately.

### 3. Update the CSV with the operator image location

Open this file
`manifests/devconsole/0.1.0/devconsole-operator.v0.1.0.clusterserviceversion.yaml` and change the image to point to the location pushed in the previous step.

Inside the file look for `image: REPLACE_IMAGE` and specify the image location.

### 4. Build the operator registry image

Now you are going to build the operator image using `test/olm/Dockerfile.registry`

```
docker build -f test/olm/Dockerfile.registry . -t quay.io/<username>/operator-registry:0.1.0 \
	--build-arg image=quay.io/<username>/devconsole-operator --build-arg version=0.1.0
docker push quay.io/<username>/operator-registry:0.1.0
```

When running the above command, substitute the `username` with your quay.io username.

### 5. Create CatalogSource and Subscription

Use this template to create a YAML file, say `cat-sub.yaml`:

```
apiVersion: operators.coreos.com/v1alpha1
kind: CatalogSource
metadata:
  name: my-catalog
  namespace: olm
spec:
  sourceType: grpc
  image: quay.io/<username>/operator-registry:0.1.0
  displayName: Community Operators
  publisher: Red Hat
---
apiVersion: operators.coreos.com/v1alpha1
kind: Subscription
metadata:
  name: my-devconsole
  namespace: operators
spec:
  channel: alpha
  name: devconsole
  source: my-catalog
  sourceNamespace: olm
```

Before applying the above file, point to the newly created operator registry image (substitute the `username` with your quay.io username).

Example:

```
oc apply -f cat-sub.yaml
```

### 6. Verify gitsources CRD presence

Check for the existence of a Custom Resource Definitions with the name as `gitsources.devconsole.openshift.io`

Run this command to list CRDs:

```
oc get crds
```
