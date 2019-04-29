#!/bin/bash
set +x

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
sh ./devconsole.sh
sh ./create_user.sh
