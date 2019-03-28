# It's necessary to set this because some environments don't link sh -> bash.
SHELL := /bin/bash

include ./make/verbose.mk

.DEFAULT_GOAL := help
include ./make/help.mk

# Create output directory for artifacts and test results. ./out is supposed to
# be a safe place for all targets to write to while knowing that all content
# inside of ./out is wiped once "make clean" is run.
$(shell mkdir -p ./out);

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