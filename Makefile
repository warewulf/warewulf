.PHONY: all

# auto installed tooling
TOOLS_DIR := .tools
TOOLS_BIN := $(TOOLS_DIR)/bin

# tools
GO_TOOLS_BIN := $(addprefix $(TOOLS_BIN)/, $(notdir $(GO_TOOLS)))
GO_TOOLS_VENDOR := $(addprefix vendor/, $(GO_TOOLS))
GOLANGCI_LINT := $(TOOLS_BIN)/golangci-lint
GOLANGCI_LINT_VERSION := v1.31.0

# built tags needed for wwbuild binary
WW_BUILD_GO_BUILD_TAGS := containers_image_openpgp containers_image_ostree

# set the go tools into the tools bin.
setup_tools: $(GO_TOOLS_BIN) $(GOLANGCI_LINT)

# install go tools into TOOLS_BIN
$(GO_TOOLS_BIN):
	@GOBIN="$(PWD)/$(TOOLS_BIN)" go install -mod=vendor $(GO_TOOLS)

# install golangci-lint into TOOLS_BIN
$(GOLANGCI_LINT):
	@curl -qq -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(TOOLS_BIN) $(GOLANGCI_LINT_VERSION)


setup: vendor $(TOOLS_DIR) setup_tools

# vendor
vendor:
	@go mod tidy -v
	@go mod vendor

$(TOOLS_DIR):
	@mkdir -p $@

# Lint
lint:
	@echo Running golangci-lint...
	@$(GOLANGCI_LINT) run --build-tags "$(WW_BUILD_GO_BUILD_TAGS)" ./...

all: warewulfd wwctl wwclient

files: all
	install -d -m 0755 /var/warewulf/
	install -d -m 0755 /etc/warewulf/
	install -d -m 0755 /etc/warewulf/ipxe
	install -d -m 0755 /var/lib/tftpboot/warewulf/ipxe/
	install -m 0640 etc/dhcpd.conf /etc/dhcp/dhcpd.conf
	install -m 0644 etc/nodes.conf /etc/warewulf/nodes.conf
	install -m 0644 etc/warewulf.conf /etc/warewulf/warewulf.conf
	install -m 0644 etc/ipxe/default.ipxe /etc/warewulf/ipxe/default.ipxe 
	cp -r tftpboot/* /var/lib/tftpboot/warewulf/ipxe/
	restorecon -r /var/lib/tftpboot/warewulf
	cp -r overlays /var/warewulf/
	chmod +x /var/warewulf/overlays/system/default/init
	mkdir -p /var/warewulf/overlays/system/default/warewulf/bin/
	cp wwclient /var/warewulf/overlays/system/default/warewulf/bin/

services: files
	sudo systemctl enable tftp
	sudo systemctl restart tftp
	sudo systemctl enable dhcpd
	sudo systemctl restart dhcpd

warewulfd:
	cd cmd/warewulfd; go build -mod vendor -o ../../warewulfd

wwctl:
	cd cmd/wwctl; go build -mod vendor -tags "$(WW_BUILD_GO_BUILD_TAGS)" -o ../../wwctl

wwclient:
	cd cmd/wwclient; CGO_ENABLED=0 GOOS=linux go build -mod vendor -a -ldflags '-extldflags -static' -o ../../wwclient

clean:
	rm -f warewulfd
	rm -f wwbuild
	rm -f wwctl

install: files services
