TOOLS_DIR := .tools
TOOLS_BIN := $(TOOLS_DIR)/bin

GO_TOOLS_BIN := $(addprefix $(TOOLS_BIN)/, $(notdir $(GO_TOOLS)))
GO_TOOLS_VENDOR := $(addprefix vendor/, $(GO_TOOLS))

GOLANGCI_LINT := $(TOOLS_BIN)/golangci-lint
GOLANGCI_LINT_VERSION := v2.11.3

GOLANG_DEADCODE := $(TOOLS_BIN)/deadcode

GOLANG_LICENSES := $(TOOLS_BIN)/go-licenses

GOLANG_STATICCHECK := $(TOOLS_BIN)/staticcheck

SHELLCHECK := $(TOOLS_BIN)/shellcheck
SHELLCHECK_VERSION := v0.11.0
SHELLCHECK_OS := $(shell uname -s | tr '[:upper:]' '[:lower:]')
SHELLCHECK_ARCH := $(shell uname -m)

$(TOOLS_DIR):
	mkdir -p $@

.PHONY: tools
tools: $(GO_TOOLS_BIN) $(GOLANGCI_LINT) $(SHELLCHECK)

$(GO_TOOLS_BIN):
	GOBIN="$(PWD)/$(TOOLS_BIN)" go install -mod=vendor $(GO_TOOLS)

$(GOLANGCI_LINT):
	curl -qq -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(TOOLS_BIN) $(GOLANGCI_LINT_VERSION)

$(GOLANG_DEADCODE):
	GOBIN="$(PWD)/$(TOOLS_BIN)" go install golang.org/x/tools/cmd/deadcode@v0.34.0

$(GOLANG_LICENSES):
	GOBIN="$(PWD)/$(TOOLS_BIN)" go install github.com/google/go-licenses@v1.6.0

$(GOLANG_STATICCHECK):
	GOBIN="$(PWD)/$(TOOLS_BIN)" go install honnef.co/go/tools/cmd/staticcheck@v0.7.0

$(SHELLCHECK):
	mkdir -p $(TOOLS_BIN)
	curl -qq -sSfL https://github.com/koalaman/shellcheck/releases/download/$(SHELLCHECK_VERSION)/shellcheck-$(SHELLCHECK_VERSION).$(SHELLCHECK_OS).$(SHELLCHECK_ARCH).tar.xz \
	    | tar -xJ -C $(TOOLS_BIN) --strip-components=1 shellcheck-$(SHELLCHECK_VERSION)/shellcheck

.PHONY: cleantools
cleantools:
	rm -rf $(TOOLS_DIR)
