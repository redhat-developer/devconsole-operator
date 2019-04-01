# It's necessary to set this because some environments don't link sh -> bash.
SHELL := /bin/bash

include ./make/verbose.mk
.DEFAULT_GOAL := help
include ./make/help.mk
include ./make/out.mk
include ./make/find-tools.mk
include ./make/go.mk
include ./make/git.mk
include ./make/dev.mk
include ./make/format.mk
include ./make/lint.mk
include ./make/deploy.mk
include ./make/test.mk
include ./make/docker.mk

.PHONY: build
## Build the operator
build: prebuild-check deps check-go-format
	@echo "building $(BINARY_SERVER_BIN)..."
	operator-sdk generate k8s
	go build -v -o $(BINARY_SERVER_BIN) cmd/manager/main.go

# -------------------------------------------------------------------
# deploy
# -------------------------------------------------------------------

LOCAL_TEST_NAMESPACE ?= "local-test"
APP_SERVICE_IMAGE_NAME?="redhat-developer/app-service"
.PHONY: local
## Run Operator locally
local: deploy-rbac build deploy-crd
	@-oc new-project $(LOCAL_TEST_NAMESPACE)
	operator-sdk up local --namespace=$(LOCAL_TEST_NAMESPACE)

.PHONY: deploy-rbac
## Setup service account and deploy RBAC
deploy-rbac:
	@-oc login -u system:admin
	@-oc create -f deploy/service_account.yaml
	@-oc create -f deploy/role.yaml
	@-oc create -f deploy/role_binding.yaml

.PHONY: deploy-crd
## Deploy CRD
deploy-crd:
	@-oc apply -f deploy/crds/devopsconsole_v1alpha1_component_crd.yaml
	@-oc apply -f deploy/crds/devopsconsole_v1alpha1_gitsource_crd.yaml
	@-oc apply -f deploy/crds/devopsconsole_v1alpha1_installer_crd.yaml

.PHONY: deploy-operator
## Deploy Operator
deploy-operator: deploy-crd
	oc create -f deploy/operator.yaml

.PHONY: deploy-clean
## Deploy a CR as test
deploy-clean:
	@-oc delete project $(LOCAL_TEST_NAMESPACE)

.PHONY: deploy-test
## Deploy a CR as test
deploy-test:
	@-oc new-project $(LOCAL_TEST_NAMESPACE)
	@-oc apply -f examples/devopsconsole_v1alpha1_component_cr.yaml
build: ./out/operator

.PHONY: clean
clean:
	$(Q)-rm -rf ${V_FLAG} ./out
	$(Q)-rm -rf ${V_FLAG} ./vendor
	$(Q)go clean ${X_FLAG} ./...

./vendor: Gopkg.toml Gopkg.lock
	$(Q)dep ensure ${V_FLAG} -vendor-only

./out/operator: ./vendor $(shell find . -path ./vendor -prune -o -name '*.go' -print)
	#$(Q)operator-sdk generate k8s
	$(Q)CGO_ENABLED=0 GOARCH=amd64 GOOS=linux \
		go build ${V_FLAG} \
		-ldflags "-X ${GO_PACKAGE_PATH}/cmd/manager.Commit=${GIT_COMMIT_ID} -X ${GO_PACKAGE_PATH}/cmd/manager.BuildTime=${BUILD_TIME}" \
		-o ./out/operator \
		cmd/manager/main.go
