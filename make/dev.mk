ifndef DEV_MK
DEV_MK:=# Prevent repeated "-include".

include ./make/verbose.mk
include ./make/git.mk

DOCKER_REPO?=quay.io/openshiftio
IMAGE_NAME?=devconsole-operator

DEVOPSCONSOLE_OPERATOR_IMAGE?=quay.io/redhat-developers/devopsconsole-operator
TIMESTAMP:=$(shell date +%s)
TAG?=$(GIT_COMMIT_ID_SHORT)-$(TIMESTAMP)

.PHONY: create-resources
create-resources:
	$(info Logging using system:admin...)
	$(Q)oc login -u system:admin
	$(info Creating sub resources...)
	$(info Creating CRDs...)
	$(Q)oc create -f ./deploy/crds/devopsconsole_v1alpha1_gitsource_crd.yaml
	$(info Creating Namespace)
	$(Q)oc create -f ./deploy/namespace.yaml
	$(info oc project codeready-devconsole)
	$(Q)oc project codeready-devconsole
	$(info Creating Service Account)
	$(Q)oc create -f ./deploy/service_account.yaml
	$(info Creating Role)
	$(Q)oc create -f ./deploy/role.yaml
	$(info Creating RoleBinding)
	$(Q)oc create -f ./deploy/role_binding.yaml

.PHONY: create-cr
create-cr:
	$(info Creating Custom Resource...)

.PHONY: build-operator-image
## Build and create the operator container image
build-operator-image:
	operator-sdk build $(DEVOPSCONSOLE_OPERATOR_IMAGE)

.PHONY: push-operator-image
## Push the operator container image to a container registry
push-operator-image: build-operator-image
	@docker login -u $(QUAY_USERNAME) -p $(QUAY_PASSWORD) $(REGISTRY_URI)
	docker push $(DEVOPSCONSOLE_OPERATOR_IMAGE)

.PHONY: deploy-operator-only
deploy-operator-only:
	@echo "Creating Deployment for Operator"
	@cat minishift/operator.yaml | sed s/\:dev/:$(TAG)/ | oc create -f -

.PHONY: clean-all
clean-all:  clean-operator clean-resources

.PHONY: clean-operator
clean-operator:
	@echo "Deleting Deployment for Operator"
	@cat minishift/operator.yaml | sed s/\:dev/:$(TAG)/ | oc delete -f - || true

.PHONY: clean-resources
clean-resources:
	@echo "Deleting sub resources..."
	@echo "Deleting ClusterRoleBinding"
	@oc delete -f ./deploy/role_binding.yaml || true
	@echo "Deleting ClusterRole"
	@oc delete -f ./deploy/role.yaml || true
	@echo "Deleting Service Account"
	@oc delete -f ./deploy/service_account.yaml || true
	@echo "Deleting Custom Resource Definitions..."
	@oc delete -f ./deploy/crds/devopsconsole_v1alpha1_gitsource_crd.yaml || true

.PHONY: deploy-operator
deploy-operator: build build-image deploy-operator-only

.PHONY: minishift-start
minishift-start:
	minishift start --cpus 4 --memory 8GB
	-eval `minishift docker-env` && oc login -u system:admin

.PHONY: deploy-all
deploy-all: clean-resources create-resources deps prebuild-check deploy-operator create-cr

endif
