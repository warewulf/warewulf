TOOLS_DIR := .tools
TOOLS_BIN := $(TOOLS_DIR)/bin

GO_TOOLS_BIN := $(addprefix $(TOOLS_BIN)/, $(notdir $(GO_TOOLS)))
GO_TOOLS_VENDOR := $(addprefix vendor/, $(GO_TOOLS))

GOLANGCI_LINT := $(TOOLS_BIN)/golangci-lint
GOLANGCI_LINT_VERSION := v1.63.4

GOLANG_DEADCODE := $(TOOLS_BIN)/deadcode

GOLANG_LICENSES := $(TOOLS_BIN)/go-licenses

GOLANG_STATICCHECK := $(TOOLS_BIN)/staticcheck

PROTOC := $(TOOLS_BIN)/protoc
PROTOC_GEN_GO := $(TOOLS_BIN)/protoc-gen-go
PROTOC_GEN_GO_GRPC := $(TOOLS_BIN)/protoc-gen-go-grpc
PROTOC_GEN_GRPC_GATEWAY := $(TOOLS_BIN)/protoc-gen-grpc-gateway

ifeq ($(ARCH),aarch64)
PROTOC_URL := https://github.com/protocolbuffers/protobuf/releases/download/v24.0/protoc-24.0-linux-aarch_64.zip
PROTOC_GEN_GRPC_GATEWAY_URL := https://github.com/grpc-ecosystem/grpc-gateway/releases/download/v2.26.0/protoc-gen-grpc-gateway-v2.26.0-linux-arm64
else
PROTOC_URL := https://github.com/protocolbuffers/protobuf/releases/download/v24.0/protoc-24.0-linux-x86_64.zip
PROTOC_GEN_GRPC_GATEWAY_URL := https://github.com/grpc-ecosystem/grpc-gateway/releases/download/v2.26.0/protoc-gen-grpc-gateway-v2.26.0-linux-x86_64
endif

$(TOOLS_DIR):
	mkdir -p $@

.PHONY: tools
tools: $(GO_TOOLS_BIN) $(GOLANGCI_LINT) $(PROTOC_GEN_GO) $(PROTOC_GEN_GO_GRPC)

$(GO_TOOLS_BIN):
	GOBIN="$(PWD)/$(TOOLS_BIN)" go install -mod=vendor $(GO_TOOLS)

$(GOLANGCI_LINT):
	curl -qq -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(TOOLS_BIN) $(GOLANGCI_LINT_VERSION)

$(GOLANG_DEADCODE):
	GOBIN="$(PWD)/$(TOOLS_BIN)" go install golang.org/x/tools/cmd/deadcode@v0.24.0

$(PROTOC): $(TOOLS_DIR)
	cd $(TOOLS_DIR) && curl -LO $(PROTOC_URL) && unzip -o $(notdir $(PROTOC_URL))
	touch --no-create $(PROTOC) # by default the timestamp is preserved from the archive

$(PROTOC_GEN_GRPC_GATEWAY):
	curl -L $(PROTOC_GEN_GRPC_GATEWAY_URL) -o $(PROTOC_GEN_GRPC_GATEWAY)
	chmod +x $(PROTOC_GEN_GRPC_GATEWAY)

$(PROTOC_GEN_GO):
	GOBIN="$(PWD)/$(TOOLS_BIN)" go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.36.5

$(PROTOC_GEN_GO_GRPC):
	GOBIN="$(PWD)/$(TOOLS_BIN)" go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.5.1

$(GOLANG_LICENSES):
	GOBIN="$(PWD)/$(TOOLS_BIN)" go install github.com/google/go-licenses@v1.6.0

$(GOLANG_STATICCHECK):
	GOBIN="$(PWD)/$(TOOLS_BIN)" go install honnef.co/go/tools/cmd/staticcheck@v0.5.1

.PHONY: cleantools
cleantools:
	rm -rf $(TOOLS_DIR)
