ifndef LINT_MK
LINT_MK:=# Prevent repeated "-include".

GOLANGCI_LINT_BIN=./out/golangci-lint

include ./make/verbose.mk
include ./make/go.mk

.PHONY: lint
## Runs linters on Go code files and YAML files
lint: lint-go-code lint-yaml courier

YAML_FILES := $(shell find . -type f -regex ".*y[a]ml" | grep -v vendor)
.PHONY: lint-yaml
## runs yamllint on all yaml files
lint-yaml: ./vendor ${YAML_FILES}
	$(Q)yamllint -c .yamllint $^

.PHONY: lint-go-code
## Checks the code with golangci-lint
lint-go-code: ./vendor $(GOLANGCI_LINT_BIN)
	# This is required for OpenShift CI enviroment
	# Ref: https://github.com/openshift/release/pull/3438#issuecomment-482053250
	$(Q)GOCACHE=$(shell pwd)/out/gocache ./out/golangci-lint ${V_FLAG} run

$(GOLANGCI_LINT_BIN):
	$(Q)curl -sfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b ./out v1.16.0

.PHONY: courier
## Validate manifests using operator-courier
courier:
	$(Q)python3 -m venv ./out/venv3
	$(Q)./out/venv3/bin/pip install --upgrade setuptools
	$(Q)./out/venv3/bin/pip install --upgrade pip
	$(Q)./out/venv3/bin/pip install operator-courier==1.3.0
	$(eval package_yaml := ./manifests/devconsole/devconsole.package.yaml)
	$(eval devconsole_version := $(shell cat $(package_yaml) | grep "currentCSV"| cut -d "." -f2- | cut -d "v" -f2 | tr -d '[:space:]'))
	$(Q)cp ./deploy/crds/*.yaml ./manifests/devconsole/$(devconsole_version)/
	# flatten command is throwing error. suppress it for now
	@-./out/venv3/bin/operator-courier flatten ./manifests/devconsole ./out/manifests-flat
	$(Q)./out/venv3/bin/operator-courier verify ./out/manifests-flat

endif

