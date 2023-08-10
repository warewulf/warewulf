include Variables.mk

.PHONY: all
all: config vendor wwctl wwclient man_pages wwapid wwapic wwapird

.PHONY: build
build: lint test-it vet all

.PHONY: setup_tools
setup_tools: $(GO_TOOLS_BIN) $(GOLANGCI_LINT)

$(GO_TOOLS_BIN):
	GOBIN="$(PWD)/$(TOOLS_BIN)" go install -mod=vendor $(GO_TOOLS)

$(GOLANGCI_LINT):
	curl -qq -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(TOOLS_BIN) $(GOLANGCI_LINT_VERSION)

.PHONY: setup
setup: vendor $(TOOLS_DIR) setup_tools

vendor:
ifndef OFFLINE_BUILD
	  go mod tidy -v
	  go mod vendor
endif

$(TOOLS_DIR):
	mkdir -p $@

.PHONY: config
config: etc/wwapic.conf \
	etc/wwapid.conf \
	etc/wwapird.conf \
	include/systemd/warewulfd.service \
	internal/pkg/config/buildconfig.go \
	warewulf.spec

%: %.in
	sed -ne "$(foreach V,$(VARLIST),s,@$V@,$(strip $($V)),g;)p" $@.in >$@

.PHONY: lint
lint: setup_tools
	$(GOLANGCI_LINT) run --build-tags "$(WW_GO_BUILD_TAGS)" --skip-dirs internal/pkg/staticfiles ./...

.PHONY: vet
vet:
	go vet ./...

.PHONY: test-it
test-it:
	go test -v ./...

.PHONY: test-cover
test-cover:
	rm -fr coverage
	mkdir coverage
	go list -f '{{if gt (len .TestGoFiles) 0}}"go test -covermode count -coverprofile {{.Name}}.coverprofile -coverpkg ./... {{.ImportPath}}"{{end}}' ./... | xargs -I {} bash -c {}
	echo "mode: count" > coverage/cover.out
	grep -h -v "^mode:" *.coverprofile >> "coverage/cover.out"
	rm *.coverprofile
	go tool cover -html=coverage/cover.out -o=coverage/cover.html

.PHONY: debian
debian: all

.PHONY: install
install: all
	install -d -m 0755 $(DESTDIR)$(BINDIR)
	install -d -m 0755 $(DESTDIR)$(WWCHROOTDIR)
	install -d -m 0755 $(DESTDIR)$(WWPROVISIONDIR)
	install -d -m 0755 $(DESTDIR)$(WWOVERLAYDIR)/wwinit/$(WWCLIENTDIR)
	install -d -m 0755 $(DESTDIR)$(WWCONFIGDIR)/ipxe
	install -d -m 0755 $(DESTDIR)$(BASHCOMPDIR)
	install -d -m 0755 $(DESTDIR)$(MANDIR)/man1
	install -d -m 0755 $(DESTDIR)$(MANDIR)/man5
	install -d -m 0755 $(DESTDIR)$(WWDOCDIR)
	install -d -m 0755 $(DESTDIR)$(FIREWALLDDIR)
	install -d -m 0755 $(DESTDIR)$(SYSTEMDDIR)
	install -d -m 0755 $(DESTDIR)$(WWDATADIR)/ipxe
	test -f $(DESTDIR)$(WWCONFIGDIR)/warewulf.conf || install -m 644 etc/warewulf.conf $(DESTDIR)$(WWCONFIGDIR)
	test -f $(DESTDIR)$(WWCONFIGDIR)/nodes.conf || install -m 644 etc/nodes.conf $(DESTDIR)$(WWCONFIGDIR)
	test -f $(DESTDIR)$(WWCONFIGDIR)/wwapic.conf || install -m 644 etc/wwapic.conf $(DESTDIR)$(WWCONFIGDIR)
	test -f $(DESTDIR)$(WWCONFIGDIR)/wwapid.conf || install -m 644 etc/wwapid.conf $(DESTDIR)$(WWCONFIGDIR)
	test -f $(DESTDIR)$(WWCONFIGDIR)/wwapird.conf || install -m 644 etc/wwapird.conf $(DESTDIR)$(WWCONFIGDIR)
	test -f $(DESTDIR)$(DATADIR)/warewulf/defaults.conf || ./wwctl --emptyconf genconfig defaults > $(DESTDIR)$(DATADIR)/warewulf/defaults.conf
	cp -r etc/examples $(DESTDIR)$(WWCONFIGDIR)/
	cp -r etc/ipxe $(DESTDIR)$(WWCONFIGDIR)/
	cp -r overlays/* $(DESTDIR)$(WWOVERLAYDIR)/
	chmod 755 $(DESTDIR)$(WWOVERLAYDIR)/wwinit/init
	find $(DESTDIR)$(WWOVERLAYDIR) -type f -name "*.in" -exec rm -f {} \;
	chmod 755 $(DESTDIR)$(WWOVERLAYDIR)/wwinit/$(WWCLIENTDIR)/wwinit
	chmod 600 $(DESTDIR)$(WWOVERLAYDIR)/wwinit/etc/ssh/ssh*
	chmod 600 $(DESTDIR)$(WWOVERLAYDIR)/wwinit/etc/NetworkManager/system-connections/ww4-managed.ww
	chmod 644 $(DESTDIR)$(WWOVERLAYDIR)/wwinit/etc/ssh/ssh*.pub.ww
	chmod 600 $(DESTDIR)$(WWOVERLAYDIR)/wwinit/warewulf/config.ww
	chmod 750 $(DESTDIR)$(WWOVERLAYDIR)/host
	install -m 0755 wwctl $(DESTDIR)$(BINDIR)
	install -m 0755 wwclient $(DESTDIR)$(WWOVERLAYDIR)/wwinit/$(WWCLIENTDIR)/wwclient
	install -m 0755 wwapic $(DESTDIR)$(BINDIR)
	install -m 0755 wwapid $(DESTDIR)$(BINDIR)
	install -m 0755 wwapird $(DESTDIR)$(BINDIR)
	install -m 0644 include/firewalld/warewulf.xml $(DESTDIR)$(FIREWALLDDIR)
	install -m 0644 include/systemd/warewulfd.service $(DESTDIR)$(SYSTEMDDIR)
	install -m 0644 LICENSE.md $(DESTDIR)$(WWDOCDIR)
	./wwctl --warewulfconf etc/warewulf.conf genconfig completions > $(DESTDIR)$(BASHCOMPDIR)/wwctl
	cp man_pages/*.1* $(DESTDIR)$(MANDIR)/man1/
	cp man_pages/*.5* $(DESTDIR)$(MANDIR)/man5/
	install -m 0644 staticfiles/README-ipxe.md $(DESTDIR)$(WWDATADIR)/ipxe
	install -m 0644 staticfiles/arm64.efi $(DESTDIR)$(WWDATADIR)/ipxe
	install -m 0644 staticfiles/x86_64.efi $(DESTDIR)$(WWDATADIR)/ipxe
	install -m 0644 staticfiles/x86_64.kpxe $(DESTDIR)$(WWDATADIR)/ipxe

.PHONY: init
init:
	systemctl daemon-reload
	cp -r tftpboot/* $(WWTFTPDIR)/ipxe/
	restorecon -r $(WWTFTPDIR)

wwctl: config vendor $(WWCTL_DEPS)
	cd cmd/wwctl; GOOS=linux go build -mod vendor -tags "$(WW_GO_BUILD_TAGS)" \
	-o ../../wwctl

wwclient: config vendor $(WWCLIENT_DEPS)
	cd cmd/wwclient; CGO_ENABLED=0 GOOS=linux go build -mod vendor -a -ldflags "-extldflags -static" \
	-o ../../wwclient

man_pages: wwctl
	install -d man_pages
	./wwctl --emptyconf genconfig man man_pages 
	cp docs/man/man5/*.5 ./man_pages/
	cd man_pages; for i in wwctl*1 *.5; do gzip --force $$i; echo -n "$$i "; done; echo

update_configuration: vendor cmd/update_configuration/update_configuration.go
	cd cmd/update_configuration && go build \
	 -X 'github.com/hpcng/warewulf/internal/pkg/node.ConfigFile=./etc/nodes.conf'"\
	 -mod vendor -tags "$(WW_GO_BUILD_TAGS)" -o ../../update_configuration

.PHONY: dist
dist: vendor config
	rm -rf .dist/$(WAREWULF)-$(VERSION) $(WAREWULF)-$(VERSION).tar.gz
	mkdir -p .dist/$(WAREWULF)-$(VERSION)
	rsync -a --exclude=".*" --exclude "*~" * .dist/$(WAREWULF)-$(VERSION)/
	cd .dist; tar -czf ../$(WAREWULF)-$(VERSION).tar.gz $(WAREWULF)-$(VERSION)
	rm -rf .dist

.PHONY: reference
reference: wwctl
	mkdir -p userdocs/reference
	./wwctl --emptyconf genconfig reference userdocs/reference/

latexpdf: reference
	make -C userdocs latexpdf

# wwapi generate code from protobuf. Requires protoc and protoc-grpc-gen-gateway to generate code.
# To setup latest protoc:
#    Download the protobuf-all-[VERSION].tar.gz from https://github.com/protocolbuffers/protobuf/releases
#    Extract the contents and change in the directory
#    ./configure
#    make
#    make check
#    sudo make install
#    sudo ldconfig # refresh shared library cache.
# To setup protoc-gen-grpc-gateway, see https://github.com/grpc-ecosystem/grpc-gateway
.PHONY: proto
proto: 
	protoc -I /usr/include -I internal/pkg/api/routes/v1 -I=. \
		--grpc-gateway_out=. \
		--grpc-gateway_opt logtostderr=true \
		--go_out=. \
		--go-grpc_out=. \
		routes.proto

wwapid:
	go build -o ./wwapid internal/app/api/wwapid/wwapid.go

wwapic:
	go build -o ./wwapic  internal/app/api/wwapic/wwapic.go

wwapird:
	go build -o ./wwapird internal/app/api/wwapird/wwapird.go

.PHONY: contclean
contclean:
	rm -f Defaults.mk
	rm -f $(WAREWULF)-$(VERSION).tar.gz
	rm -f bash_completion
	rm -f config_defaults
	rm -f etc/wwapi{c,d,rd}.conf
	rm -f etc/wwapi{c,d,rd}.config
	rm -f include/systemd/warewulfd.service
	rm -f internal/pkg/buildconfig/setconfigs.go
	rm -f internal/pkg/config/buildconfig.go
	rm -f man_page 
	rm -f print_defaults 
	rm -f update_configuration
	rm -f usr/share/man/man1/
	rm -f warewulf.spec
	rm -f warewulf-*.tar.gz
	rm -f wwapic 
	rm -f wwapid 
	rm -f wwapird
	rm -f wwclient
	rm -f wwctl
	rm -rf $(TOOLS_DIR)
	rm -rf bash_completion.d
	rm -rf /config
	rm -rf .dist/
	rm -rf _dist/
	rm -rf etc/bash_completion.d/
	rm -rf man_pages
	rm -rf userdocs/_*
	rm -rf userdocs/reference/*

.PHONY: clean
clean: contclean
	rm -rf vendor

.PHONY: debinstall
debinstall: files debfiles
