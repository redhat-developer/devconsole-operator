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

.PHONY: test-e2e
## Runs the e2e tests without coverage
test-e2e: build build-image e2e-setup
	$(Q)go test ./test/e2e/... -root=$(PWD) -kubeconfig=$(HOME)/.kube/config -globalMan deploy/test/global-manifests.yaml -namespacedMan deploy/test/namespace-manifests.yaml ${V_FLAG} -parallel=1 -singleNamespace

.PHONY: e2e-setup
## TODO: TBD
e2e-setup: e2e-cleanup
	$(Q)oc new-project devconsole-e2e-test || true

.PHONY: e2e-cleanup
## TODO: TBD
e2e-cleanup:
	$(Q)oc login -u system:admin
	$(Q)oc delete project devconsole-e2e-test || true
