#!/bin/bash
set -x
OC_LOGIN_USERNAME=kubeadmin
oc login -u ${OC_LOGIN_USERNAME} -p ${OC_LOGIN_PASSWORD}

oc apply -f ./yamls/unmanage.yaml
oc scale --replicas 0 deployment console-operator --namespace openshift-console-operator
sleep 20s
oc scale --replicas 0 deployment console --namespace openshift-console
sleep 20s
oc apply -f ./yamls/redeploy-console-operator.yaml
sleep 25s
oc scale --replicas 1 deployment console --namespace openshift-console
sleep 10s
sh ./devconsole.sh
sh ./create_user.sh
