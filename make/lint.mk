ifndef LINT_MK
LINT_MK:=# Prevent repeated "-include".

include ./make/verbose.mk
include ./make/go.mk

.PHONY: lint
## Runs linters on Go code files and YAML files
lint: lint-go-code lint-yaml

YAML_FILES := $(shell find . -type f -regex ".*y[a]ml" | grep -v vendor)
.PHONY: lint-yaml
## runs yamllint on all yaml files
lint-yaml: ${YAML_FILES}
	$(Q)yamllint -c .yamllint $^

.PHONY: lint-go-code
## Checks the code with golangci-lint
lint-go-code:
	$(Q)go get github.com/golangci/golangci-lint/cmd/golangci-lint
	$(Q)${GOPATH}/bin/golangci-lint ${V_FLAG} run

endif
