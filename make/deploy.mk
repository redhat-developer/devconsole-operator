ifndef DEPLOY_MK
DEPLOY_MK:=# Prevent repeated "-include".

include ./make/verbose.mk

# to watch all namespaces, keep namespace empty
APP_NAMESPACE ?= ""
LOCAL_TEST_NAMESPACE ?= "local-test"

.PHONY: local
## Run Operator locally
local: deploy-rbac build deploy-crd
	$(Q)-oc new-project $(LOCAL_TEST_NAMESPACE)
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
	$(Q)-oc apply -f deploy/crds/devconsole_v1alpha1_component_crd.yaml
	$(Q)-oc apply -f deploy/crds/devconsole_v1alpha1_gitsource_crd.yaml
	$(Q)-oc apply -f deploy/crds/devconsole_v1alpha1_gitsourceanalysis_crd.yaml

.PHONY: deploy-operator
## Deploy Operator
deploy-operator: deploy-crd
	$(Q)oc create -f deploy/operator.yaml

.PHONY: deploy-clean
## Deploy a CR as test
deploy-clean:
	$(Q)-oc delete project $(LOCAL_TEST_NAMESPACE)

.PHONY: deploy-test
## Deploy a CR as test
deploy-test:
deploy-test:
	$(Q)-oc new-project $(LOCAL_TEST_NAMESPACE)
	$(Q)-oc apply -f examples/devconsole_v1alpha1_gitsource_cr.yaml
	$(Q)-oc apply -f examples/devconsole_v1alpha1_component_cr.yaml
	$(Q)-oc apply -f examples/devconsole_v1alpha1_gitsourceanalysis_cr.yaml

endif

