# -------------------------------------------------------------------
# deploy
# -------------------------------------------------------------------

# to watch all namespaces, keep namespace empty
APP_NAMESPACE ?= ""
.PHONY: local
## Run Operator locally
local: deploy-rbac build deploy-crd
	$(Q)-oc new-project $(APP_NAMESPACE)
	$(Q)operator-sdk up local --namespace=$(APP_NAMESPACE)

.PHONY: deploy-rbac
## Setup service account and deploy RBAC
deploy-rbac:
	$(Q)-oc login -u system:admin
	$(Q)-oc create -f deploy/service_account.yaml
	$(Q)-oc create -f deploy/role.yaml
	$(Q)-oc create -f deploy/role_binding.yaml

.PHONY: deploy-crd
## Deploy CRD
deploy-crd:
	$(Q)-oc apply -f deploy/crds/devopsconsole_v1alpha1_component_crd.yaml
	$(Q)-oc apply -f deploy/crds/devopsconsole_v1alpha1_gitsource_crd.yaml

.PHONY: deploy-operator
## Deploy Operator
deploy-operator: deploy-crd
	$(Q)oc create -f deploy/operator.yaml

.PHONY: deploy-clean
## Deploy a CR as test
deploy-clean:
	$(Q)-oc delete component.devopsconsole.openshift.io/myapp
	$(Q)-oc delete imagestream.image.openshift.io/myapp-builder
	$(Q)-oc delete imagestream.image.openshift.io/myapp-output
	$(Q)-oc delete buildconfig.build.openshift.io/myapp-bc
	$(Q)-oc delete deploymentconfig.apps.openshift.io/myapp

.PHONY: deploy-test
## Deploy a CR as test
deploy-test:
	$(Q)oc create -f examples/devopsconsole_v1alpha1_component_cr.yaml
