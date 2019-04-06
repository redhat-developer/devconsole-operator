ifndef TEST_MK
TEST_MK:=# Prevent repeated "-include".
UNAME_S := $(shell uname -s)

include ./make/verbose.mk
include ./make/out.mk

export DEPLOYED_NAMESAPCE:=

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
	$(Q)-oc delete -f ./deploy/crds/devconsole_v1alpha1_gitsource_crd.yaml
	$(Q)-oc delete -f ./deploy/service_account.yaml --namespace $(TEST_NAMESPACE)
	$(Q)-oc delete -f ./deploy/role.yaml --namespace $(TEST_NAMESPACE)
	$(Q)-oc delete -f ./deploy/test/role_binding_test.yaml --namespace $(TEST_NAMESPACE)
	$(Q)-oc delete -f ./deploy/test/operator_test.yaml --namespace $(TEST_NAMESPACE)

.PHONY: test-olm-integration
## Runs the OLM integration tests without coverage
test-olm-integration: push-operator-image olm-integration-setup
	$(eval DEPLOYED_NAMESAPCE := operators)
	$(call log-info,"Running OLM integration test: $@")
	$(Q)oc apply -f https://raw.githubusercontent.com/operator-framework/operator-lifecycle-manager/master/deploy/upstream/quickstart/olm.yaml
	$(eval package_yaml := ./manifests/devconsole/devconsole.package.yaml)
	$(eval devconsole_version := $(shell cat $(package_yaml) | grep "currentCSV"| cut -d "." -f2- | cut -d "v" -f2 | tr -d '[:space:]'))
	$(Q)docker build -f Dockerfile.registry . -t $(DEVCONSOLE_OPERATOR_REGISTRY_IMAGE):$(devconsole_version)-$(TAG) \
		--build-arg image=$(DEVCONSOLE_OPERATOR_IMAGE):$(TAG) --build-arg version=$(devconsole_version)
	@docker login -u $(QUAY_USERNAME) -p $(QUAY_PASSWORD) $(REGISTRY_URI)
	$(Q)docker push $(DEVCONSOLE_OPERATOR_REGISTRY_IMAGE):$(devconsole_version)-$(TAG)

	$(Q)sed -e "s,REPLACE_IMAGE,$(DEVCONSOLE_OPERATOR_REGISTRY_IMAGE):$(devconsole_version)-$(TAG)," ./test/e2e/catalog_source.yaml | oc apply -f -
	$(Q)oc apply -f ./test/e2e/subscription.yaml

	$(Q)operator-sdk test local ./test/e2e/ --go-test-flags "-v -parallel=1"

.PHONY: olm-integration-setup
olm-integration-setup: olm-integration-cleanup
	$(Q)oc new-project $(TEST_NAMESPACE)

.PHONY: olm-integration-cleanup
olm-integration-cleanup: get-test-namespace
	$(Q)oc login -u system:admin
	$(Q)-oc delete subscription my-devconsole -n operators
	$(Q)-oc delete catalogsource my-catalog -n olm
	# The following cleanup is required due to a potential bug in the test framework.
	$(Q)-oc delete clusterroles.rbac.authorization.k8s.io "devconsole-operator"
	$(Q)-oc delete clusterrolebindinhttps://github.com/redhat-developer/devconsole-operator/pull/54gs.rbac.authorization.k8s.io "devconsole-operator"
	$(Q)-oc delete project $(TEST_NAMESPACE)  --wait

#-------------------------------------------------------------------------------
# e2e test in dev mode
#-------------------------------------------------------------------------------

.PHONY: e2e-cleanup
## Create a namespace used in e2e tests
e2e-cleanup: get-test-namespace
	$(Q)-oc login -u system:admin
	$(Q)-oc delete -f ./deploy/crds/devconsole_v1alpha1_component_crd.yaml
	$(Q)-oc delete -f ./deploy/service_account.yaml --namespace $(TEST_NAMESPACE)
	$(Q)-oc delete -f ./deploy/role.yaml --namespace $(TEST_NAMESPACE)
	$(Q)-oc delete -f ./deploy/test/role_binding_test.yaml --namespace $(TEST_NAMESPACE)
	$(Q)-oc delete -f ./deploy/test/operator_test.yaml --namespace $(TEST_NAMESPACE)

.PHONY: e2e-setup
## Create a namespace used in e2e tests
e2e-setup: e2e-cleanup
	$(Q)-oc new-project $(TEST_NAMESPACE)

.PHONY: build-image-local
build-image-local: e2e-setup
	eval $$(minishift docker-env) && operator-sdk build $(shell minishift openshift registry)/$(TEST_NAMESPACE)/devconsole-operator

.PHONY: test-e2e-local
test-e2e-local: build-image-local
	$(eval DEPLOYED_NAMESAPCE := $(TEST_NAMESPACE))
	$(Q)-oc login -u system:admin
	$(Q)-oc project $(TEST_NAMESPACE)
	$(Q)-oc create -f ./deploy/crds/devconsole_v1alpha1_component_crd.yaml
	$(Q)-oc create -f ./deploy/crds/devconsole_v1alpha1_gitsource_crd.yaml
	$(Q)-oc create -f ./deploy/service_account.yaml --namespace $(TEST_NAMESPACE)
	$(Q)-oc create -f ./deploy/role.yaml --namespace $(TEST_NAMESPACE)
ifeq ($(UNAME_S),Darwin	)
	$(Q)sed -i "" 's|REPLACE_NAMESPACE|$(TEST_NAMESPACE)|g' ./deploy/test/role_binding_test.yaml
else
	$(Q)sed -i 's|REPLACE_NAMESPACE|$(TEST_NAMESPACE)|g' ./deploy/test/role_binding_test.yaml
endif
	@-oc create -f ./deploy/test/role_binding_test.yaml --namespace $(TEST_NAMESPACE)
ifeq ($(UNAME_S),Darwin)
	$(Q)sed -i "" 's|REPLACE_IMAGE|172.30.1.1:5000/$(TEST_NAMESPACE)/devconsole-operator:latest|g' ./deploy/test/operator_test.yaml
else
	$(Q)sed -i 's|REPLACE_IMAGE|172.30.1.1:5000/$(TEST_NAMESPACE)/devconsole-operator:latest|g' ./deploy/test/operator_test.yaml
endif
	@eval $$(minishift docker-env) && oc create -f ./deploy/test/operator_test.yaml --namespace $(TEST_NAMESPACE)
ifeq ($(UNAME_S),Darwin)
	$(Q)sed -i "" 's|$(TEST_NAMESPACE)|REPLACE_NAMESPACE|g' ./deploy/test/role_binding_test.yaml
	$(Q)sed -i "" 's|172.30.1.1:5000/$(TEST_NAMESPACE)/devconsole-operator:latest|REPLACE_IMAGE|g' ./deploy/test/operator_test.yaml
else
	$(Q)sed -i 's|$(TEST_NAMESPACE)|REPLACE_NAMESPACE|g' ./deploy/test/role_binding_test.yaml
	$(Q)sed -i 's|172.30.1.1:5000/$(TEST_NAMESPACE)/devconsole-operator:latest|REPLACE_IMAGE|g' ./deploy/test/operator_test.yaml
endif
	$(Q)eval $$(minishift docker-env) && operator-sdk test local ./test/e2e --namespace $(TEST_NAMESPACE) --no-setup
endif
