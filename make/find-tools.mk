# Check all required tools are accessible
REQUIRED_EXECUTABLES = go gofmt dep git operator-sdk oc sed
ifeq ($(MAKECMDGOALS),docker-image)
    REQUIRED_EXECUTABLES = docker
endif
ifeq ($(MAKECMDGOALS),docker-run)
    REQUIRED_EXECUTABLES = docker
endif
ifneq ($(MAKECMDGOALS),help)
ifneq ($(MAKECMDGOALS),)
ifeq ($(VERBOSE),1)
$(info Searching for required executables: $(REQUIRED_EXECUTABLES)...)
endif
K := $(foreach exec,$(REQUIRED_EXECUTABLES),\
        $(if $(shell which $(exec) 2>/dev/null),some string,$(error "ERROR: No "$(exec)" binary found in in PATH!")))
endif
endif
