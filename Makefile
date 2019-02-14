PROJECT_NAME=devopsconsole-operator
PACKAGE_NAME:=github.com/fabric8-services/$(PROJECT_NAME)
CUR_DIR=$(shell pwd)
TMP_PATH=$(CUR_DIR)/tmp
INSTALL_PREFIX=$(CUR_DIR)/bin
VENDOR_DIR=vendor
INCLUDE_DIR=make
SOURCE_DIR ?= .
SOURCES := $(shell find $(SOURCE_DIR) -path $(SOURCE_DIR)/vendor -prune -o -name '*.go' -print)

ifeq ($(OS),Windows_NT)
	include ./$(INCLUDE_DIR)/Makefile.win
else
	include ./$(INCLUDE_DIR)/Makefile.lnx
endif

# declares variable that are OS-sensitive
include ./$(INCLUDE_DIR)/test.mk
include ./$(INCLUDE_DIR)/Makefile.dev

DOCKER_BIN := $(shell command -v $(DOCKER_BIN_NAME) 2> /dev/null)

ifneq ($(OS),Windows_NT)
	ifdef DOCKER_BIN
		include ./$(INCLUDE_DIR)/docker.mk
	endif
endif

# This is a fix for a non-existing user in passwd file when running in a docker
# container and trying to clone repos of dependencies
GIT_COMMITTER_NAME ?= "user"
GIT_COMMITTER_EMAIL ?= "user@example.com"
export GIT_COMMITTER_NAME
export GIT_COMMITTER_EMAIL

COMMIT=$(shell git rev-parse HEAD 2>/dev/null)
GITUNTRACKEDCHANGES := $(shell git status --porcelain --untracked-files=no)
ifneq ($(GITUNTRACKEDCHANGES),)
	COMMIT := $(COMMIT)-dirty
endif
BUILD_TIME=`date -u '+%Y-%m-%dT%H:%M:%SZ'`

.DEFAULT_GOAL := help

# Call this function with $(call log-info,"Your message")
define log-info =
	@echo "INFO: $(1)"
endef

# -------------------------------------------------------------------
# Docker build
# -------------------------------------------------------------------
BUILD_DIR = bin
REGISTRY_URI = quay.io
REGISTRY_NS = ${PROJECT_NAME}
REGISTRY_IMAGE = ${PROJECT_NAME}

ifeq ($(TARGET),rhel)
	REGISTRY_URL := ${REGISTRY_URI}/openshiftio/rhel-${REGISTRY_NS}-${REGISTRY_IMAGE}
	DOCKERFILE := Dockerfile.rhel
else
	REGISTRY_URL := ${REGISTRY_URI}/openshiftio/${REGISTRY_NS}-${REGISTRY_IMAGE}
	DOCKERFILE := Dockerfile
endif

$(BUILD_DIR):
	mkdir $(BUILD_DIR)

.PHONY: build-linux $(BUILD_DIR)
build-linux: makefiles prebuild-check deps ## Builds the Linux binary for the container image into bin/ folder
	CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build -v $(LDFLAGS) -o $(BUILD_DIR)/$(PROJECT_NAME)

image: clean-artifacts build-linux
	docker build -t $(REGISTRY_URL) \
	  --build-arg BUILD_DIR=$(BUILD_DIR)\
	  --build-arg PROJECT_NAME=$(PROJECT_NAME)\
	  -f $(VENDOR_DIR)/github.com/fabric8-services/fabric8-common/makefile/$(DOCKERFILE) .

# -------------------------------------------------------------------
# help!
# -------------------------------------------------------------------

.PHONY: help
# Based on https://gist.github.com/rcmachado/af3db315e31383502660
## Display this help text.
help:/
	$(info Available targets)
	$(info -----------------)
	@awk '/^[a-zA-Z\-\_0-9]+:/ { \
		helpMessage = match(lastLine, /^## (.*)/); \
		helpCommand = substr($$1, 0, index($$1, ":")-1); \
		if (helpMessage) { \
			helpMessage = substr(lastLine, RSTART + 3, RLENGTH); \
			gsub(/##/, "\n                                     ", helpMessage); \
			printf "%-35s - %s\n", helpCommand, helpMessage; \
			lastLine = "" \
		} \
	} \
	{ hasComment = match(lastLine, /^## (.*)/); \
          if(hasComment) { \
            lastLine=lastLine$$0; \
	  } \
          else { \
	    lastLine = $$0 \
          } \
        }' $(MAKEFILE_LIST)
# -------------------------------------------------------------------
# required tools
# -------------------------------------------------------------------

# Find all required tools:
GIT_BIN := $(shell command -v $(GIT_BIN_NAME) 2> /dev/null)
TMP_BIN_DIR := $(TMP_PATH)/bin
DEP_BIN := $(TMP_BIN_DIR)/$(DEP_BIN_NAME)
DEP_VERSION=v0.5.0
GOLANGCI_LINT_VERSION=1.12
GOLANGCI_LINT_BIN=$(TMP_BIN_DIR)/$(GOLANGCI_LINT_BIN_NAME)
GO_BIN := $(shell command -v $(GO_BIN_NAME) 2> /dev/null)

$(INSTALL_PREFIX):
	mkdir -p $(INSTALL_PREFIX)
$(TMP_PATH):
	mkdir -p $(TMP_PATH)

.PHONY: prebuild-check
prebuild-check: $(TMP_PATH) $(INSTALL_PREFIX)
# Check that all tools where found
ifndef GIT_BIN
	$(error The "$(GIT_BIN_NAME)" executable could not be found in your PATH)
endif
ifndef DEP_BIN
	$(error The "$(DEP_BIN_NAME)" executable could not be found in your PATH)
endif
ifndef GO_BIN
	$(error The "$(GO_BIN_NAME)" executable could not be found in your PATH)
endif

# -------------------------------------------------------------------
# deps
# -------------------------------------------------------------------

.PHONY: deps
## Download build dependencies.
deps: $(DEP_BIN) $(VENDOR_DIR)

# install dep in a the tmp/bin dir of the repo
$(DEP_BIN):
	@echo "Installing 'dep' $(DEP_VERSION) at '$(TMP_BIN_DIR)'..."
	mkdir -p $(TMP_BIN_DIR)
ifeq ($(UNAME_S),Darwin)
	@curl -L -s https://github.com/golang/dep/releases/download/$(DEP_VERSION)/dep-darwin-amd64 -o $(DEP_BIN)
	@cd $(TMP_BIN_DIR) && \
	curl -L -s https://github.com/golang/dep/releases/download/$(DEP_VERSION)/dep-darwin-amd64.sha256 -o $(TMP_BIN_DIR)/dep-darwin-amd64.sha256 && \
	echo "1a7bdb0d6c31ecba8b3fd213a1170adf707657123e89dff234871af9e0498be2  dep" > dep-darwin-amd64.sha256 && \
	shasum -a 256 --check dep-darwin-amd64.sha256
else
	@curl -L -s https://github.com/golang/dep/releases/download/$(DEP_VERSION)/dep-linux-amd64 -o $(DEP_BIN)
	@cd $(TMP_BIN_DIR) && \
	echo "287b08291e14f1fae8ba44374b26a2b12eb941af3497ed0ca649253e21ba2f83  dep" > dep-linux-amd64.sha256 && \
	sha256sum -c dep-linux-amd64.sha256
endif
	@chmod +x $(DEP_BIN)

$(VENDOR_DIR): Gopkg.toml
	@echo "checking dependencies with $(DEP_BIN_NAME)"
	@$(DEP_BIN) ensure -v

# -------------------------------------------------------------------
# Code format/check
# -------------------------------------------------------------------
GOFORMAT_FILES := $(shell find  . -name '*.go' | grep -vEf $(INCLUDE_DIR)/gofmt_exclude)

.PHONY: check-go-format
# Exists with an error if there are files whose formatting differs from gofmt's
check-go-format: prebuild-check deps ## Exists with an error if there are files whose formatting differs from gofmt's
	@gofmt -s -l ${GOFORMAT_FILES} 2>&1 \
		| tee /tmp/gofmt-errors \
		| read \
	&& echo "ERROR: These files differ from gofmt's style (run 'make format-go-code' to fix this):" \
	&& cat /tmp/gofmt-errors \
	&& exit 1 \
	|| true

.PHONY: analyze-go-code
# Run golangci analysis over the code.
analyze-go-code: deps ## Run golangci analysis over the code.
	$(info >>--- RESULTS: GOLANGCI CODE ANALYSIS ---<<)
	@go get -u github.com/golangci/golangci-lint/cmd/golangci-lint
	@golangci-lint run

.PHONY: format-go-code
# Formats any go file that differs from gofmt's style
format-go-code: prebuild-check ## Formats any go file that differs from gofmt's style
	@gofmt -s -l -w ${GOFORMAT_FILES}

# -------------------------------------------------------------------
# Code format/check with golangci-lint
# -------------------------------------------------------------------
.PHONY: check-go-code
## Checks the code with golangci-lint (see .golangci.yaml)
check-go-code: $(GOLANGCI_LINT_BIN)
	@echo "checking code..."
	$(GOLANGCI_LINT_BIN) run

# install golangci-lint in the 'tmp/bin' dir of the repo
$(GOLANGCI_LINT_BIN):
	@echo "Installing 'golang-ci-lint' $(GOLANGCI_LINT_VERSION) at '$(TMP_BIN_DIR)' ..."
	mkdir -p $(TMP_BIN_DIR)
ifeq ($(UNAME_S),Darwin)
	@curl -L -s https://github.com/golangci/golangci-lint/releases/download/v$(GOLANGCI_LINT_VERSION)/golangci-lint-$(GOLANGCI_LINT_VERSION)-darwin-amd64.tar.gz -o $(GOLANGCI_LINT_BIN)-$(GOLANGCI_LINT_VERSION)-darwin-amd64.tar.gz && \
	cd $(TMP_BIN_DIR) && curl -L -s https://github.com/golangci/golangci-lint/releases/download/v$(GOLANGCI_LINT_VERSION)/golangci-lint-$(GOLANGCI_LINT_VERSION)-checksums.txt -o  golangci-lint-$(GOLANGCI_LINT_VERSION)-checksums.txt && \
	grep "darwin-amd64" golangci-lint-$(GOLANGCI_LINT_VERSION)-checksums.txt > golangci-lint-$(GOLANGCI_LINT_VERSION)-checksum-darwin-amd64.txt && \
	shasum -a 256 --check golangci-lint-$(GOLANGCI_LINT_VERSION)-checksum-darwin-amd64.txt && \
	tar xvf $(GOLANGCI_LINT_BIN)-$(GOLANGCI_LINT_VERSION)-darwin-amd64.tar.gz -C $(TMP_BIN_DIR) && \
	mv $(TMP_BIN_DIR)/golangci-lint-$(GOLANGCI_LINT_VERSION)-darwin-amd64/golangci-lint $(GOLANGCI_LINT_BIN)
else
	@curl -L -s https://github.com/golangci/golangci-lint/releases/download/v$(GOLANGCI_LINT_VERSION)/golangci-lint-$(GOLANGCI_LINT_VERSION)-linux-amd64.tar.gz -o $(GOLANGCI_LINT_BIN)-$(GOLANGCI_LINT_VERSION)-linux-amd64.tar.gz && \
	cd $(TMP_BIN_DIR) && curl -L -s https://github.com/golangci/golangci-lint/releases/download/v$(GOLANGCI_LINT_VERSION)/golangci-lint-$(GOLANGCI_LINT_VERSION)-checksums.txt -o  golangci-lint-$(GOLANGCI_LINT_VERSION)-checksums.txt && \
	grep "linux-amd64" golangci-lint-$(GOLANGCI_LINT_VERSION)-checksums.txt > golangci-lint-$(GOLANGCI_LINT_VERSION)-checksum-linux-amd64.txt && \
	sha256sum -c golangci-lint-$(GOLANGCI_LINT_VERSION)-checksum-linux-amd64.txt && \
	tar xvf $(GOLANGCI_LINT_BIN)-$(GOLANGCI_LINT_VERSION)-linux-amd64.tar.gz -C $(TMP_BIN_DIR) && \
	mv $(TMP_BIN_DIR)/golangci-lint-$(GOLANGCI_LINT_VERSION)-linux-amd64/golangci-lint $(GOLANGCI_LINT_BIN)
endif
	@chmod +x $(GOLANGCI_LINT_BIN)


# -------------------------------------------------------------------
# clean
# -------------------------------------------------------------------

# For the global "clean" target all targets in this variable will be executed
CLEAN_TARGETS =

CLEAN_TARGETS += clean-artifacts
.PHONY: clean-artifacts
# Removes the ./bin directory.
clean-artifacts:
	-rm -rf $(INSTALL_PREFIX)

CLEAN_TARGETS += clean-object-files
.PHONY: clean-object-files
# Runs go clean to remove any executables or other object files.
clean-object-files:
	go clean ./...

CLEAN_TARGETS += clean-vendor
.PHONY: clean-vendor
# Removes the ./vendor directory.
clean-vendor:
	-rm -rf $(VENDOR_DIR)

CLEAN_TARGETS += clean-tmp
.PHONY: clean-tmp
# Removes the ./vendor directory.
clean-tmp:
	-rm -rf $(TMP_PATH)

# Keep this "clean" target here after all `clean-*` sub tasks
.PHONY: clean
## Runs all clean-* targets.
clean: $(CLEAN_TARGETS)

# -------------------------------------------------------------------
# build the binary executable (to ship in prod)
# -------------------------------------------------------------------

.PHONY: build
## Build the server
build: prebuild-check deps
	@echo "building $(BINARY_SERVER_BIN)..."
	go build -v -o $(BINARY_SERVER_BIN) cmd/manager/main.go
