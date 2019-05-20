#!/bin/bash
set +x

cd $(dirname $(readlink -f $0))

oc apply -f ./yamls/unmanage.yaml
oc scale --replicas 0 deployment console-operator --namespace openshift-console-operator
oc scale --replicas 0 deployment console --namespace openshift-console
oc apply -f ./yamls/redeploy-console-operator.yaml
#It takes time to get the pod in running state
while [ "$(oc get pods --field-selector=status.phase=Running -n openshift-console-operator)" == "No resources found." ]
do
    sleep 1s
done

oc scale --replicas 1 deployment console --namespace openshift-console
#Delete the already existing pod in the openshift-console namespace
#Because it's a Deployment, Kubernetes will automatically recreate the pod and pull the latest image.
#Have also updated the image pull policy to Always in the yamls/redeploy-console-operator.yaml

CONSOLE_POD="$(oc get pods -o=name -n openshift-console | cut -d'/' -f2- | cut -f 1 -d "-" | head -n 1)"
CONSOLE_POD_NAME="$(oc get pods -o=name -n openshift-console | cut -d'/' -f2- | cut -d'-' -f1- | head -n 1)"
if echo "${CONSOLE_POD}" == "console";then
    oc delete pod ${CONSOLE_POD_NAME} -n openshift-console
fi
sh ./devconsole.sh
sh ./create_user.sh
