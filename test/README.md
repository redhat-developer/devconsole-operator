## E2E Tests

### How to Run

#### Test e2e in dev mode

```
make minishift-start
eval $(minishift docker-env)
make test-e2e-local
```

## Steps to verify operator registry on OCP4
### Pre-requisites
* have a 48h-ephemeral cluster on AWS for OCP4
* have a quay account
* `oc` client installed
> NOTE: you can also do all those steps with the UI

### 1. Build and push image to quay
```
QUAY_USERNAME=YYY QUAY_PASSWORD=XXX REGISTRY_ORG=YYY make push-operator-image
```
> NOTE: to test you can push to your own username in quay, therefore QUAY_USERNAME and REGISTRY_ORG are the same

### 2. create a catalog 
This catalog contains a link to the operator image you've just created in step1.
* Create a new file `catalog.yaml`
```
apiVersion: operators.coreos.com/v1alpha1
kind: CatalogSource
metadata:
  name: my-catalog
  namespace: openshift-operator-lifecycle-manager
spec:
  sourceType: grpc
  image: quay.io/YYY/operator-registry:0.1.0-2613b51-dirty-1554395857
  displayName: Community Operators
  publisher: Red Hat
```
* login to OCP4 
```
oc apply -f catalog.yaml
oc get catsrc —all-namespaces
```
You should see your new catalog.

### 3. create a subscription
* Create a new file `subscription.yaml`
```
apiVersion: operators.coreos.com/v1alpha1
kind: Subscription
metadata:
  generateName: devconsole-
  namespace: openshift-operators
spec:
  source: my-catalog
  sourceNamespace: openshift-operator-lifecycle-manager
  name: devconsole
  startingCSV: devconsole-operator.v0.1.0
  channel: alpha
```

### 4. Create a new Component
```
oc new-project demo
oc apply -f examples/devconsole_v1alpha1_gitsource_cr.yaml
oc apply -f examples/devconsole_v1alpha1_component_cr.yaml
```
You should be able to see the route of your nodejs app.

## Steps to verify operator registry on minishift

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

Now you are going to build the operator image using `Dockerfile.registry`

```
docker build -f Dockerfile.registry . -t quay.io/<username>/operator-registry:0.1.0 \
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
