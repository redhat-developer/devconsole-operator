ifndef TEST_MK
TEST_MK:=# Prevent repeated "-include".

include ./make/verbose.mk
include ./make/out.mk

.PHONY: test
## Runs Go package tests and stops when the first one fails
test: ./vendor
	$(Q)go test -vet off ${V_FLAG} $(shell go list ./... | grep -v /test/e2e) -failfast

.PHONY: test-coverage
## Runs Go package tests and produces coverage information
test-coverage: ./out/cover.out

.PHONY: test-coverage-html
## Gather (if needed) coverage information and show it in your browser
test-coverage-html: ./vendor ./out/cover.out
	$(Q)go tool cover -html=./out/cover.out

./out/cover.out: ./vendor
	$(Q)go test ${V_FLAG} -race $(shell go list ./... | grep -v /test/e2e) -failfast -coverprofile=cover.out -covermode=atomic -outputdir=./out

.PHONY: get-test-namespace
get-test-namespace: ./out/test-namespace
	$(eval TEST_NAMESPACE := $(shell cat ./out/test-namespace))

./out/test-namespace:
	@echo -n "test-namespace-$(shell uuidgen | tr '[:upper:]' '[:lower:]')" > ./out/test-namespace

.PHONY: test-e2e
## Runs the e2e tests without coverage
test-e2e: build build-image e2e-setup
	$(info Running E2E test: $@)
	$(Q)go test ./test/e2e/... \
		-parallel=1 \
		${V_FLAG} \
		-root=$(PWD) \
		-kubeconfig=$(HOME)/.kube/config \
		-globalMan ./deploy/test/global-manifests.yaml \
		-namespacedMan ./deploy/test/namespace-manifests.yaml \
		-singleNamespace

.PHONY: e2e-setup
## TODO: TBD
e2e-setup: e2e-cleanup 
	$(Q)-oc new-project $(TEST_NAMESPACE)

.PHONY: e2e-cleanup
## TODO: TBD
e2e-cleanup: get-test-namespace
	$(Q)-oc login -u system:admin
	$(Q)-oc delete -f ./deploy/crds/devconsole_v1alpha1_component_crd.yaml
	$(Q)-oc delete -f ./deploy/service_account.yaml --namespace $(TEST_NAMESPACE)
	$(Q)-oc delete -f ./deploy/role.yaml --namespace $(TEST_NAMESPACE)
	$(Q)-oc delete -f ./deploy/test/role_binding_test.yaml --namespace $(TEST_NAMESPACE)
	$(Q)-oc delete -f ./deploy/test/operator_test.yaml --namespace $(TEST_NAMESPACE)

#-------------------------------------------------------------------------------
# e2e test in dev mode
#-------------------------------------------------------------------------------

.PHONY: build-image-local
build-image-local: e2e-setup
	eval $$(minishift docker-env) && operator-sdk build $(shell minishift openshift registry)/$(TEST_NAMESPACE)/devconsole-operator

.PHONY: e2e-local
e2e-local: build-image-local
	$(Q)-oc login -u system:admin
	$(Q)-oc project $(TEST_NAMESPACE)
	$(Q)-oc create -f ./deploy/crds/devconsole_v1alpha1_component_crd.yaml
	$(Q)-oc create -f ./deploy/service_account.yaml --namespace $(TEST_NAMESPACE)
	$(Q)-oc create -f ./deploy/role.yaml --namespace $(TEST_NAMESPACE)
ifeq ($(UNAME_S),Darwin)
	$(Q)sed ${QUIET_FLAG} -i "" 's|REPLACE_NAMESPACE|$(TEST_NAMESPACE)|g' ./deploy/test/role_binding_test.yaml
else
	$(Q)sed ${QUIET_FLAG} -i 's|REPLACE_NAMESPACE|$(TEST_NAMESPACE)|g' ./deploy/test/role_binding_test.yaml
endif
	$(Q)-oc create -f ./deploy/test/role_binding_test.yaml --namespace $(TEST_NAMESPACE)
ifeq ($(UNAME_S),Darwin)
	$(Q)sed ${QUIET_FLAG} -i "" 's|REPLACE_IMAGE|172.30.1.1:5000/$(TEST_NAMESPACE)/devconsole-operator:latest|g' ./deploy/test/operator_test.yaml
else
	$(Q)sed ${QUIET_FLAG} -i 's|REPLACE_IMAGE|172.30.1.1:5000/$(TEST_NAMESPACE)/devconsole-operator:latest|g' ./deploy/test/operator_test.yaml
endif
	@eval $$(minishift docker-env) && oc create -f ./deploy/test/operator_test.yaml --namespace $(TEST_NAMESPACE)
ifeq ($(UNAME_S),Darwin)
	$(Q)sed ${QUIET_FLAG} -i "" 's|$(TEST_NAMESPACE)|REPLACE_NAMESPACE|g' ./deploy/test/role_binding_test.yaml
	$(Q)sed ${QUIET_FLAG} -i "" 's|172.30.1.1:5000/$(TEST_NAMESPACE)/devconsole-operator:latest|REPLACE_IMAGE|g' ./deploy/test/operator_test.yaml
else
	$(Q)sed ${QUIET_FLAG} -i 's|$(TEST_NAMESPACE)|REPLACE_NAMESPACE|g' ./deploy/test/role_binding_test.yaml
	$(Q)sed ${QUIET_FLAG} -i 's|172.30.1.1:5000/$(TEST_NAMESPACE)/devconsole-operator:latest|REPLACE_IMAGE|g' ./deploy/test/operator_test.yaml
endif
	$(Q)eval $$(minishift docker-env) && operator-sdk test local ./test/e2e --namespace $(TEST_NAMESPACE) --no-setup

endif
