.PHONY: all

VERSION ?= 4.2.0
RELEASE ?= 1

SRC ?= HEAD

VERSION_FULL ?= $(shell test -e .git && git describe --tags --long --first-parent --always)
ifeq ($(VERSION_FULL),)
VERSION_FULL := $(VERSION)
endif

# System locations
PREFIX ?= /usr/local
BINDIR ?= $(PREFIX)/bin
SYSCONFDIR ?= $(PREFIX)/etc
SRVDIR ?= $(PREFIX)/srv
SHAREDIR ?= $(PREFIX)/share
MANDIR ?= $(SHAREDIR)/man
LOCALSTATEDIR ?= $(PREFIX)/var

TFTPDIR ?= /var/lib/tftpboot
FIREWALLDDIR ?= /usr/lib/firewalld/services
SYSTEMDDIR ?= /usr/lib/systemd/system
BASH_COMPLETION ?= /etc/bash_completion.d/

# Warewulf locations
WWPROVISIONDIR ?= $(SRVDIR)/warewulf
WWOVERLAYDIR ?= $(LOCALSTATEDIR)/warewulf/overlays
WWCHROOTDIR ?= $(LOCALSTATEDIR)/warewulf/chroots

# SuSE
#TFTPDIR ?= /srv/tftpboot
#FIREWALLDIR ?= /srv/tftp

# auto installed tooling
TOOLS_DIR := .tools
TOOLS_BIN := $(TOOLS_DIR)/bin
CONFIG := $(shell pwd)

# tools
GO_TOOLS_BIN := $(addprefix $(TOOLS_BIN)/, $(notdir $(GO_TOOLS)))
GO_TOOLS_VENDOR := $(addprefix vendor/, $(GO_TOOLS))
GOLANGCI_LINT := $(TOOLS_BIN)/golangci-lint
GOLANGCI_LINT_VERSION := v1.31.0

# use GOPROXY for older git clients and speed up downloads
GOPROXY ?= https://proxy.golang.org
export GOPROXY

# built tags needed for wwbuild binary
WW_BUILD_GO_BUILD_TAGS := containers_image_openpgp containers_image_ostree

all: config vendor wwctl wwclient bash_completion.d man_pages

build: lint test-it vet all

# set the go tools into the tools bin.
setup_tools: $(GO_TOOLS_BIN) $(GOLANGCI_LINT)

# install go tools into TOOLS_BIN
$(GO_TOOLS_BIN):
	@GOBIN="$(PWD)/$(TOOLS_BIN)" go install -mod=vendor $(GO_TOOLS)

# install golangci-lint into TOOLS_BIN
$(GOLANGCI_LINT):
	@curl -qq -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(TOOLS_BIN) $(GOLANGCI_LINT_VERSION)


setup: vendor $(TOOLS_DIR) setup_tools config

# vendor
vendor:
	go mod tidy -v
	go mod vendor

$(TOOLS_DIR):
	@mkdir -p $@

# Pre-build steps for source, such as "go generate"
config:
	set -x ;\
	for i in `find . -type f -name "*.in" -not -path "./vendor/*"`; do \
		NAME=`echo $$i | sed -e 's,\.in,,'`; \
		sed -e 's,@BINDIR@,$(BINDIR),g' \
			-e 's,@SYSCONFDIR@,$(SYSCONFDIR),g' \
			-e 's,@LOCALSTATEDIR@,$(LOCALSTATEDIR),g' \
			-e 's,@SRVDIR@,$(SRVDIR),g' \
			-e 's,@TFTPDIR@,$(TFTPDIR),g' \
			-e 's,@FIREWALLDDIR@,$(FIREWALLDDIR),g' \
			-e 's,@SYSTEMDDIR@,$(SYSTEMDDIR),g' \
			-e 's,@WWOVERLAYDIR@,$(WWOVERLAYDIR),g' \
			-e 's,@WWCHROOTDIR@,$(WWCHROOTDIR),g' \
			-e 's,@WWPROVISIONDIR@,$(WWPROVISIONDIR),g' \
			-e 's,@VERSION@,$(VERSION),g' \
			-e 's,@RELEASE@,$(RELEASE),g' $$i > $$NAME; \
	done
	touch config

# Lint
lint: setup_tools
	@echo Running golangci-lint...
	@$(GOLANGCI_LINT) run --build-tags "$(WW_BUILD_GO_BUILD_TAGS)" --skip-dirs internal/pkg/staticfiles ./...

vet:
	go vet ./...

test-it:
	go test -v ./...

# Generate test coverage
test-cover:     ## Run test coverage and generate html report
	rm -fr coverage
	mkdir coverage
	go list -f '{{if gt (len .TestGoFiles) 0}}"go test -covermode count -coverprofile {{.Name}}.coverprofile -coverpkg ./... {{.ImportPath}}"{{end}}' ./... | xargs -I {} bash -c {}
	echo "mode: count" > coverage/cover.out
	grep -h -v "^mode:" *.coverprofile >> "coverage/cover.out"
	rm *.coverprofile
	go tool cover -html=coverage/cover.out -o=coverage/cover.html

debian: all 

files: all
	install -d -m 0755 $(DESTDIR)$(BINDIR)
	install -d -m 0755 $(DESTDIR)$(WWCHROOTDIR)
	install -d -m 0755 $(DESTDIR)$(WWPROVISIONDIR)
	install -d -m 0755 $(DESTDIR)$(WWOVERLAYDIR)
	install -d -m 0755 $(DESTDIR)$(WWOVERLAYDIR)/wwinit/bin/
	install -d -m 0755 $(DESTDIR)$(WWOVERLAYDIR)/wwinit/warewulf/bin/
	install -d -m 0755 $(DESTDIR)$(SYSCONFDIR)/warewulf/
	install -d -m 0755 $(DESTDIR)$(SYSCONFDIR)/warewulf/ipxe
	install -d -m 0755 $(DESTDIR)$(TFTPDIR)/warewulf/ipxe/
	install -d -m 0755 $(DESTDIR)$(BASH_COMPLETION)
	install -d -m 0755 $(DESTDIR)$(SHAREDIR)/man/man1
	install -d -m 0755 $(DESTDIR)$(FIREWALLDDIR)
	install -d -m 0755 $(DESTDIR)$(SYSTEMDDIR)
	test -f $(DESTDIR)$(SYSCONFDIR)/warewulf/warewulf.conf || install -m 644 etc/warewulf.conf $(DESTDIR)$(SYSCONFDIR)/warewulf/
	test -f $(DESTDIR)$(SYSCONFDIR)/warewulf/hosts.tmpl || install -m 644 etc/hosts.tmpl $(DESTDIR)$(SYSCONFDIR)/warewulf/
	test -f $(DESTDIR)$(SYSCONFDIR)/warewulf/nodes.conf || install -m 644 etc/nodes.conf $(DESTDIR)$(SYSCONFDIR)/warewulf/
	cp -r etc/dhcp $(DESTDIR)$(SYSCONFDIR)/warewulf/
	cp -r etc/ipxe $(DESTDIR)$(SYSCONFDIR)/warewulf/
	cp -r overlays/* $(DESTDIR)$(WWOVERLAYDIR)/
	chmod 755 $(DESTDIR)$(WWOVERLAYDIR)/wwinit/init
	chmod 600 $(DESTDIR)$(WWOVERLAYDIR)/wwinit/etc/ssh/ssh*
	chmod 644 $(DESTDIR)$(WWOVERLAYDIR)/wwinit/etc/ssh/ssh*.pub.ww
	install -m 0755 wwctl $(DESTDIR)$(BINDIR)
	install -c -m 0644 include/firewalld/warewulf.xml $(DESTDIR)$(FIREWALLDDIR)
	install -c -m 0644 include/systemd/warewulfd.service $(DESTDIR)$(SYSTEMDDIR)
	cp bash_completion.d/warewulf $(DESTDIR)$(BASH_COMPLETION)
	cp man_pages/* $(DESTDIR)$(MANDIR)/man1/

init:
	systemctl daemon-reload
	cp -r tftpboot/* $(TFTPDIR)/warewulf/ipxe/
	restorecon -r $(TFTPDIR)/warewulf

# Overlay file system has changed
#debfiles: debian
#	chmod +x $(DESTDIR)$(WWROOT)/warewulf/overlays/system/debian/init
#	chmod 600 $(DESTDIR)$(WWROOT)/warewulf/overlays/system/debian/etc/ssh/ssh*
#	chmod 644 $(DESTDIR)$(WWROOT)/warewulf/overlays/system/debian/etc/ssh/ssh*.pub.ww
#	mkdir -p $(DESTDIR)$(WWROOT)/warewulf/overlays/system/debian/warewulf/bin/
#	cp wwclient $(DESTDIR)$(WWROOT)/warewulf/overlays/system/debian/warewulf/bin/

wwctl:
	cd cmd/wwctl; GOOS=linux go build -mod vendor -tags "$(WW_BUILD_GO_BUILD_TAGS)" -o ../../wwctl

wwclient:
	cd cmd/wwclient; CGO_ENABLED=0 GOOS=linux go build -mod vendor -a -ldflags '-extldflags -static' -o ../../wwclient

install_wwclient: wwclient
	install -m 0755 wwclient $(DESTDIR)$(WWOVERLAYDIR)/wwinit/bin/wwclient


bash_completion:
	cd cmd/bash_completion && go build -ldflags="-X 'github.com/hpcng/warewulf/internal/pkg/warewulfconf.ConfigFile=./etc/warewulf.conf'\
	 -X 'github.com/hpcng/warewulf/internal/pkg/node.ConfigFile=./etc/nodes.conf'"\
	 -mod vendor -tags "$(WW_BUILD_GO_BUILD_TAGS)" -o ../../bash_completion

bash_completion.d: bash_completion
	install -d -m 0755 bash_completion.d
	./bash_completion  bash_completion.d/warewulf

man_page:
	cd cmd/man_page && go build -ldflags="-X 'github.com/hpcng/warewulf/internal/pkg/warewulfconf.ConfigFile=./etc/warewulf.conf'\
	 -X 'github.com/hpcng/warewulf/internal/pkg/node.ConfigFile=./etc/nodes.conf'"\
	 -mod vendor -tags "$(WW_BUILD_GO_BUILD_TAGS)" -o ../../man_page

man_pages: man_page
	install -d man_pages
	./man_page ./man_pages
	cd man_pages; for i in wwctl*1; do echo "Compressing manpage: $$i"; gzip --force $$i; done

#config_defaults:
#	cd cmd/config_defaults && go build -ldflags="-X 'github.com/hpcng/warewulf/internal/pkg/warewulfconf.ConfigFile=./etc/warewulf.conf' \
#	 -X 'github.com/hpcng/warewulf/internal/pkg/node.ConfigFile=./etc/nodes.conf' \
#	 -X 'github.com/hpcng/warewulf/internal/pkg/warewulfconf.defaultDataStore=$(WWROOT)/warewulf'" \
#	 -mod vendor -tags "$(WW_BUILD_GO_BUILD_TAGS)" -o ../../config_defaults

dist: vendor config
	rm -rf _dist/warewulf-$(VERSION)
	mkdir -p _dist/warewulf-$(VERSION)
	git archive --format=tar $(SRC) | tar -xf - -C _dist/warewulf-$(VERSION)
	cp -r vendor _dist/warewulf-$(VERSION)/
	cp warewulf.spec _dist/warewulf-$(VERSION)/
	cd _dist; tar -czf ../warewulf-$(VERSION).tar.gz warewulf-$(VERSION)

clean:
	rm -f wwclient
	rm -f wwctl
	rm -rf _dist
	rm -f warewulf-$(VERSION).tar.gz
	rm -f bash_completion
	rm -rf bash_completion.d
	rm -f man_page
	rm -rf man_pages
	rm -rf vendor
#	rm -f config_defaults
	rm -f config

install: files install_wwclient

debinstall: files debfiles

