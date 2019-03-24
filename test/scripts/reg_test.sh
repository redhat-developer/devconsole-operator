#!/bin/bash

set -x +e

# ------
# Prerequisites - Define directories for console and devopsconsole-operator and variables

export QUAY_USERNAME=$quay_username
export QUAY_PASSWORD=$quay_password
export QUAY_REPO_NAME=$quay_repo_name
export QUAY_URL=quay.io/$QUAY_REPO_NAME/devopsconsole-operator

TEST_ROOT="$(pwd)"
CONSOLE="$(readlink -f $TEST_ROOT/console.git)"
DEVOPS_CONSOLE_OPERATOR="$(readlink -f $GOPATH/src/github.com/devopsconsole-operator)"

# ------
# Download and build minishift and preinstalled console

if [ ! -f $CONSOLE/bin/bridge ]; then
    rm -rvf $CONSOLE
    git clone --depth=1 https://github.com/talamer/console $CONSOLE

    cd $CONSOLE
    ./build.sh
fi

sleep 30

# ------
# Step 1 - Configure/start cluster and install OLM

minishift start
oc login -u system:admin

export OPENSHIFT_API="$(oc status | grep "on server http" | sed -e 's,.*on server \(http.*\),\1,g')"

oc process -f $CONSOLE/examples/console-oauth-client.yaml | oc apply -f -
oc get oauthclient console-oauth-client -o jsonpath='{.secret}' > $CONSOLE/examples/console-client-secret
oc get secrets -n default --field-selector type=kubernetes.io/service-account-token -o json | jq '.items[0].data."ca.crt"' -r | python -m base64 -d > $CONSOLE/examples/ca.crt
oc adm policy add-cluster-role-to-user cluster-admin admin

oc login -u admin -p system
source $CONSOLE/contrib/oc-environment.sh
$CONSOLE/bin/bridge &
BRIDGE_PID=$!

sleep 30

CONSOLE_STATUS=$(curl --silent -LI http://localhost:9000 | head -n 1 | cut -d ' ' -f 2)

if [ "$CONSOLE_STATUS" ne 200 ]; then
    kill $BRIDGE_PID
    echo "ERROR"
    exit 1
fi

sleep 30

oc project myproject
oc create -f https://raw.githubusercontent.com/operator-framework/operator-lifecycle-manager/master/deploy/upstream/quickstart/olm.yaml 

# ------
# Step 2 - Clone/build the Operator

git clone --depth=1 https://github.com/redhat-developer/devopsconsole-operator $DEVOPS_CONSOLE_OPERATOR
cd $DEVOPS_CONSOLE_OPERATOR
make build

# ------
# Step 3 - Build and push the operator image to quay

operator-sdk build $QUAY_URL
docker login -u $QUAY_USERNAME -p $QUAY_PASSWORD quay.io

sed -i 's/REPLACE_IMAGE/$QUAY_URL/g' manifests/devopsconsole/0.1.0/devopsconsole-operator.v0.1.0.clusterserviceversion.yaml
docker push QUAY_URL

# ------
# Step 4 - Build and push the registry  

docker build -f test/olm/Dockerfile.registry . -t quay.io/$QUAY_REPO_NAME/operator-registry:0.1.0
docker push quay.io/$QUAY_REPO_NAME/operator-registry:0.1.0

# ------
# Step 5 - Create the CatalogSource and Subscription

cat <<EOT >> cat-sub.yaml
apiVersion: operators.coreos.com/v1alpha1
kind: CatalogSource
metadata:
  name: my-catalog
  namespace: olm
spec:
  sourceType: grpc
  image: quay.io/$QUAY_REPO_NAME/operator-registry:0.1.0
  displayName: Community Operators
  publisher: Red Hat
---
apiVersion: operators.coreos.com/v1alpha1
kind: Subscription
metadata:
  name: my-devopsconsole
  namespace: operators
spec:
  channel: alpha
  name: devopsconsole
  source: my-catalog
  sourceNamespace: olm
EOT

cat cat-sub.yaml

oc apply -f cat-sub.yaml
oc get crds

minishift stop



