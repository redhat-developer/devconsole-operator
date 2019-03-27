DOCKER_REPO?=quay.io/openshiftio
IMAGE_NAME?=devconsole-operator
SHORT_COMMIT=$(shell git rev-parse --short HEAD)
ifneq ($(GITUNTRACKEDCHANGES),)
SHORT_COMMIT := $(SHORT_COMMIT)-dirty
endif

DEVOPSCONSOLE_OPERATOR_IMAGE?=quay.io/redhat-developers/devopsconsole-operator
TIMESTAMP:=$(shell date +%s)
TAG?=$(SHORT_COMMIT)-$(TIMESTAMP)

DEPLOY_DIR:=deploy

.PHONY: create-resources
create-resources:
	$(Q)echo "Logging using system:admin..."
	$(Q)oc login -u system:admin
	$(Q)echo "Creating sub resources..."
	$(Q)echo "Creating CRDs..."
	$(Q)oc create -f $(DEPLOY_DIR)/crds/devopsconsole_v1alpha1_gitsource_crd.yaml
	$(Q)echo "Creating Namespace"
	$(Q)oc create -f $(DEPLOY_DIR)/namespace.yaml
	$(Q)echo "oc project codeready-devconsole"
	$(Q)oc project codeready-devconsole
	$(Q)echo "Creating Service Account"
	$(Q)oc create -f $(DEPLOY_DIR)/service_account.yaml
	$(Q)echo "Creating Role"
	$(Q)oc create -f $(DEPLOY_DIR)/role.yaml
	$(Q)echo "Creating RoleBinding"
	$(Q)oc create -f $(DEPLOY_DIR)/role_binding.yaml

.PHONY: create-cr
create-cr:
	@echo "Creating Custom Resource..."

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
	@oc delete -f $(DEPLOY_DIR)/role_binding.yaml || true
	@echo "Deleting ClusterRole"
	@oc delete -f $(DEPLOY_DIR)/role.yaml || true
	@echo "Deleting Service Account"
	@oc delete -f $(DEPLOY_DIR)/service_account.yaml || true
	@echo "Deleting Custom Resource Definitions..."
	@oc delete -f $(DEPLOY_DIR)/crds/devopsconsole_v1alpha1_gitsource_crd.yaml || true

.PHONY: deploy-operator
deploy-operator: build build-image deploy-operator-only

.PHONY: minishift-start
minishift-start:
	minishift start --cpus 4 --memory 8GB
	-eval `minishift docker-env` && oc login -u system:admin

.PHONY: deploy-all
deploy-all: clean-resources create-resources deps prebuild-check deploy-operator create-cr
