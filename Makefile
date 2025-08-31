.DEFAULT_GOAL := help

.PHONY: all
all: build

##@ General

# https://gist.github.com/prwhite/8168133
# Maybe use https://github.com/drdv/makefile-doc in the future
help: ## Display this help
	@awk 'BEGIN {FS = ":.*##"; printf "The Warewulf Makefile\n\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)
	@echo
	@echo "Define OFFLINE_BUILD=1 to avoid requiring network access."

include Variables.mk
include Tools.mk

##@ Build

.PHONY: version
version: ## Build version
	@echo $(VERSION)

.PHONY: build
build: wwctl wwclient etc/bash_completion.d/wwctl ## Build the Warewulf binaries

.PHONY: docs
docs: man_pages reference ## Build the documentation

.PHONY: spec
spec: warewulf.spec ## Create an RPM spec file

.PHONY: dist
dist: $(config) ## Create a distributable source tarball
	$(eval TMPDIR := $(shell mktemp -d))
	mkdir -p $(TMPDIR)/$(WAREWULF)-$(VERSION)
	git ls-files >$(TMPDIR)/dist-files
	tar -c --files-from $(TMPDIR)/dist-files | tar -C $(TMPDIR)/$(WAREWULF)-$(VERSION) -x
	test -d vendor/ && cp -a vendor/ $(TMPDIR)/$(WAREWULF)-$(VERSION) || :
	scripts/get-version.sh >$(TMPDIR)/$(WAREWULF)-$(VERSION)/VERSION
	tar -C $(TMPDIR) -czf $(WAREWULF)-$(VERSION).tar.gz $(WAREWULF)-$(VERSION)
	rm -rf $(TMPDIR)

RPMDIR = $(HOME)/rpmbuild
.PHONY: rpm
rpm: spec dist ## Create an RPM package
	@mkdir -p $(RPMDIR)/{BUILD,RPMS,SOURCES,SPECS,SRPMS}
	cp $(WAREWULF)-$(VERSION).tar.gz $(RPMDIR)/SOURCES/
	rpmbuild -bb warewulf.spec

config = include/systemd/warewulfd.service \
	internal/pkg/config/buildconfig.go \
	warewulf.spec
.PHONY: config
config: $(config)

$(config): Defaults.mk

%: %.in
	sed -ne "$(foreach V,$(VARLIST),s,@$V@,$(strip $($V)),g;)p" $@.in >$@

wwctl: $(config) $(call godeps,cmd/wwctl/main.go)
	GOOS=linux go build -mod vendor -tags "$(WW_GO_BUILD_TAGS)" -o wwctl cmd/wwctl/main.go

wwclient: $(config) $(call godeps,cmd/wwclient/main.go)
	CGO_ENABLED=0 GOOS=linux go build -mod vendor -a -ldflags "-extldflags -static" -o wwclient cmd/wwclient/main.go

.PHONY: man_pages
man_pages: wwctl $(wildcard docs/man/man5/*.5)
	mkdir -p docs/man/man1
	./wwctl --emptyconf genconfig man docs/man/man1
	gzip --force docs/man/man1/*.1
	for manpage in docs/man/man5/*.5; do gzip <$${manpage} >$${manpage}.gz; done

etc/bash_completion.d/wwctl: wwctl
	mkdir -p etc/bash_completion.d/
	./wwctl --emptyconf completion bash >etc/bash_completion.d/wwctl

.PHONY: reference
reference: wwctl
	mkdir -p userdocs/reference
	./wwctl --emptyconf genconfig reference userdocs/reference/

latexpdf: reference
	SPHINXOPTS='-t pdf -D release=$(VERSION)' make -C userdocs latexpdf

##@ Development

vendor: ## Create the vendor directory (if it does not exist)
	go mod vendor

.PHONY: tidy
tidy: ## Clean up golang dependencies
	go mod tidy

.PHONY: fmt
fmt: ## Update source code formatting
	go fmt ./...

.PHONY: lint
lint: $(config) ## Run the linter
	$(GOLANGCI_LINT) run --build-tags "$(WW_GO_BUILD_TAGS)" --timeout=5m ./...

.PHONY: staticcheck
staticcheck: $(GOLANG_STATICCHECK) $(config) ## Run static code check
	$(GOLANG_STATICCHECK) ./...

.PHONY: deadcode
deadcode: $(config) ## Check for unused code
	test $$($(GOLANG_DEADCODE) -test ./... | tee /dev/stderr | wc -l) = 0

.PHONY: vet
vet: $(config) ## Check for invalid code
	go vet ./...

.PHONY: test
test: $(config) ## Run full test suite
	TZ=UTC go test ./...

.PHONY: test-cover
test-cover: $(config) ## Generate a coverage report for the test suite
	rm -rf coverage
	mkdir coverage
	go list -f '{{if gt (len .TestGoFiles) 0}}"TZ=UTC go test -covermode count -coverprofile {{.Name}}.coverprofile -coverpkg ./... {{.ImportPath}}"{{end}}' ./... | xargs -I {} bash -c {}
	echo "mode: count" >coverage/cover.out
	grep -h -v "^mode:" *.coverprofile >>"coverage/cover.out"
	rm *.coverprofile
	go tool cover -html=coverage/cover.out -o=coverage/cover.html

.PHONY: LICENSE_DEPENDENCIES.md
LICENSE_DEPENDENCIES.md: $(GOLANG_LICENSES) scripts/update-license-dependencies.sh
	rm -rf vendor
	GOLANG_LICENSES=$(GOLANG_LICENSES) scripts/update-license-dependencies.sh

.PHONY: licenses
licenses: LICENSE_DEPENDENCIES.md # Update LICENSE_DEPENDENCIES.md

.PHONY: cleanconfig
cleanconfig:
	rm -f $(config)
	rm -rf etc/bash_completion.d/

.PHONY: cleantest
cleantest:
	rm -rf *.coverprofile

.PHONY: cleandist
cleandist:
	rm -f $(WAREWULF)-$(VERSION).tar.gz
	rm -rf .dist/

.PHONY: cleanmake
cleanmake:
	rm -f Defaults.mk

.PHONY: cleanbin
cleanbin:
	rm -f wwclient
	rm -f wwctl
	rm -f update_configuration

.PHONY: cleandocs
cleandocs:
	rm -rf userdocs/_*
	rm -rf userdocs/reference/*
	rm -rf docs/man/man1
	rm -rf docs/man/man5/*.gz

.PHONY: cleanvendor
cleanvendor:
	rm -rf vendor

.PHONY: clean
clean: cleanconfig cleantest cleandist cleantools cleanmake cleanbin cleandocs ## Remove built configuration, docs, binaries, and artifacts

##@ Installation

.PHONY: install
install: build docs ## Install Warewulf from source
	install -d -m 0755 $(DESTDIR)$(BINDIR)
	install -d -m 0755 $(DESTDIR)$(WWCHROOTDIR)
	install -d -m 0755 $(DESTDIR)$(WWOVERLAYDIR)
	install -d -m 0755 $(DESTDIR)$(WWPROVISIONDIR)
	install -d -m 0755 $(DESTDIR)$(DATADIR)/warewulf/overlays/wwclient/rootfs/$(WWCLIENTDIR)
	install -d -m 0755 $(DESTDIR)$(DATADIR)/warewulf/overlays/host/rootfs/$(TFTPDIR)/warewulf/
	install -d -m 0755 $(DESTDIR)$(WWCONFIGDIR)/examples
	install -d -m 0755 $(DESTDIR)$(WWCONFIGDIR)/ipxe
	install -d -m 0755 $(DESTDIR)$(WWCONFIGDIR)/grub
	install -d -m 0755 $(DESTDIR)$(BASHCOMPDIR)
	install -d -m 0755 $(DESTDIR)$(MANDIR)/man1
	install -d -m 0755 $(DESTDIR)$(MANDIR)/man5
	install -d -m 0755 $(DESTDIR)$(WWDOCDIR)
	install -d -m 0755 $(DESTDIR)$(FIREWALLDDIR)
	install -d -m 0755 $(DESTDIR)$(LOGROTATEDIR)
	install -d -m 0755 $(DESTDIR)$(SYSTEMDDIR)
	install -d -m 0755 $(DESTDIR)$(IPXESOURCE)
	install -d -m 0755 $(DESTDIR)$(DATADIR)/warewulf
	# wwctl genconfig to get the compiled in paths to warewulf.conf
	install -d -m 0755 $(DESTDIR)$(DATADIR)/warewulf/bmc
	test -f $(DESTDIR)$(WWCONFIGDIR)/warewulf.conf || ./wwctl --warewulfconf etc/warewulf.conf genconfig warewulfconf print> $(DESTDIR)$(WWCONFIGDIR)/warewulf.conf
	test -f $(DESTDIR)$(WWCONFIGDIR)/auth.conf || install -m 0600 etc/auth.conf $(DESTDIR)$(WWCONFIGDIR)
	test -f $(DESTDIR)$(WWCONFIGDIR)/nodes.conf || install -m 0644 etc/nodes.conf $(DESTDIR)$(WWCONFIGDIR)
	for f in etc/examples/*.ww; do install -m 0644 $$f $(DESTDIR)$(WWCONFIGDIR)/examples/; done
	for f in etc/ipxe/*.ipxe; do install -m 0644 $$f $(DESTDIR)$(WWCONFIGDIR)/ipxe/; done
	for f in lib/warewulf/bmc/*.tmpl; do install -m 0644 $$f $(DESTDIR)$(DATADIR)/warewulf/bmc; done
	install -m 0644 etc/grub/grub.cfg.ww $(DESTDIR)$(WWCONFIGDIR)/grub/grub.cfg.ww
	install -m 0644 etc/grub/chainload.ww $(DESTDIR)$(DATADIR)/warewulf/overlays/host/rootfs$(TFTPDIR)/warewulf/grub.cfg.ww
	install -m 0644 etc/logrotate.d/warewulfd.conf $(DESTDIR)$(LOGROTATEDIR)/warewulfd.conf
	(cd overlays && find * -path '*/internal' -prune -o -type f -exec install -D -m 0644 {} $(DESTDIR)$(DATADIR)/warewulf/overlays/{} \;)
	(cd overlays && find * -path '*/internal' -prune -o -type d -exec mkdir -pv $(DESTDIR)$(DATADIR)/warewulf/overlays/{} \;)
	(cd overlays && find * -path '*/internal' -prune -o -type l -exec cp -av {} $(DESTDIR)$(DATADIR)/warewulf/overlays/{} \;)
	chmod 0755 $(DESTDIR)$(DATADIR)/warewulf/overlays/wwinit/rootfs/init
	chmod 0755 $(DESTDIR)$(DATADIR)/warewulf/overlays/wwinit/rootfs/$(WWCLIENTDIR)/run-init
	chmod 0755 $(DESTDIR)$(DATADIR)/warewulf/overlays/wwinit/rootfs/$(WWCLIENTDIR)/run-wwinit.d
	chmod 0600 $(DESTDIR)$(DATADIR)/warewulf/overlays/wwinit/rootfs/$(WWCLIENTDIR)/config.ww
	chmod 0600 $(DESTDIR)$(DATADIR)/warewulf/overlays/ssh.host_keys/rootfs/etc/ssh/ssh*
	chmod 0644 $(DESTDIR)$(DATADIR)/warewulf/overlays/ssh.host_keys/rootfs/etc/ssh/ssh*.pub.ww
	chmod 0600 $(DESTDIR)$(DATADIR)/warewulf/overlays/NetworkManager/rootfs/etc/NetworkManager/system-connections/ww4-managed.ww
	chmod 0750 $(DESTDIR)$(DATADIR)/warewulf/overlays/host/rootfs
	install -m 0755 wwctl $(DESTDIR)$(BINDIR)
	install -m 0755 wwclient $(DESTDIR)$(DATADIR)/warewulf/overlays/wwclient/rootfs/$(WWCLIENTDIR)/wwclient
	install -m 0644 include/firewalld/warewulf.xml $(DESTDIR)$(FIREWALLDDIR)
	install -m 0644 include/systemd/warewulfd.service $(DESTDIR)$(SYSTEMDDIR)
	install -m 0644 LICENSE.md $(DESTDIR)$(WWDOCDIR)
	install -m 0644 etc/bash_completion.d/wwctl $(DESTDIR)$(BASHCOMPDIR)/wwctl
	for f in docs/man/man1/*.1.gz; do install -m 0644 $$f $(DESTDIR)$(MANDIR)/man1/; done
	for f in docs/man/man5/*.5.gz; do install -m 0644 $$f $(DESTDIR)$(MANDIR)/man5/; done
	install -pd -m 0755 $(DESTDIR)$(DRACUTMODDIR)/50wwinit
	install -m 0644 dracut/modules.d/50wwinit/*.sh  dracut/modules.d/50wwinit/*.override $(DESTDIR)$(DRACUTMODDIR)/50wwinit

.PHONY: install-sos
install-sos:
	install -D -m 0644 include/sos/warewulf.py $(DESTDIR)$(SOSPLUGINS)/warewulf.py

.PHONY: init
init:
	systemctl daemon-reload
	cp -r tftpboot/* $(WWTFTPDIR)/ipxe/
	restorecon -r $(WWTFTPDIR)

ifndef OFFLINE_BUILD
wwctl: vendor
wwclient: vendor
update_configuration: vendor
dist: vendor

lint: $(GOLANGCI_LINT)
deadcode: $(GOLANG_DEADCODE)

clean: cleanvendor
endif
