.PHONY: all
all: build

include Variables.mk
include Tools.mk

.PHONY: build
build: wwctl wwclient wwapid wwapic wwapird etc/defaults.conf etc/bash_completion.d/wwctl

.PHONY: docs
docs: man_pages reference

vendor:
	go mod vendor

.PHONY: tidy
tidy:
	go mod tidy

.PHONY: fmt
fmt:
	go fmt ./...

config = etc/wwapic.conf \
	etc/wwapid.conf \
	etc/wwapird.conf \
	include/systemd/warewulfd.service \
	internal/pkg/config/buildconfig.go \
	warewulf.spec
.PHONY: config
config: $(config)

%: %.in
	sed -ne "$(foreach V,$(VARLIST),s,@$V@,$(strip $($V)),g;)p" $@.in >$@

wwctl: $(config) $(call godeps,cmd/wwctl/main.go)
	GOOS=linux go build -mod vendor -tags "$(WW_GO_BUILD_TAGS)" -o wwctl cmd/wwctl/main.go

wwclient: $(config) $(call godeps,cmd/wwclient/main.go)
	CGO_ENABLED=0 GOOS=linux go build -mod vendor -a -ldflags "-extldflags -static" -o wwclient cmd/wwclient/main.go

update_configuration: $(config) $(call godeps,cmd/update_configuration/update_configuration.go)
	go build -X 'github.com/warewulf/warewulf/internal/pkg/node.ConfigFile=./etc/nodes.conf'" \
	    -mod vendor -tags "$(WW_GO_BUILD_TAGS)" -o update_configuration cmd/update_configuration/update_configuration.go

wwapid: $(config) $(call godeps,internal/app/api/wwapid/wwapid.go)
	go build -o ./wwapid internal/app/api/wwapid/wwapid.go

wwapic: $(config) $(call godeps,internal/app/api/wwapic/wwapic.go)
	go build -o ./wwapic  internal/app/api/wwapic/wwapic.go

wwapird: $(config) $(call godeps,internal/app/api/wwapird/wwapird.go)
	go build -o ./wwapird internal/app/api/wwapird/wwapird.go

.PHONY: man_pages
man_pages: wwctl $(wildcard docs/man/man5/*.5)
	mkdir -p docs/man/man1
	./wwctl --emptyconf genconfig man docs/man/man1
	gzip --force docs/man/man1/*.1
	for manpage in docs/man/man5/*.5; do gzip <$${manpage} >$${manpage}.gz; done

etc/defaults.conf: wwctl
	./wwctl --emptyconf genconfig defaults >etc/defaults.conf

etc/bash_completion.d/wwctl: wwctl
	mkdir -p etc/bash_completion.d/
	./wwctl --emptyconf genconfig completions >etc/bash_completion.d/wwctl

.PHONY: lint
lint: $(config)
	$(GOLANGCI_LINT) run --build-tags "$(WW_GO_BUILD_TAGS)" --timeout=5m ./...

.PHONY: vet
vet: $(config)
	go vet ./...

.PHONY: test
test: $(config)
	go test ./...

.PHONY: test-cover
test-cover: $(config)
	rm -rf coverage
	mkdir coverage
	go list -f '{{if gt (len .TestGoFiles) 0}}"go test -covermode count -coverprofile {{.Name}}.coverprofile -coverpkg ./... {{.ImportPath}}"{{end}}' ./... | xargs -I {} bash -c {}
	echo "mode: count" >coverage/cover.out
	grep -h -v "^mode:" *.coverprofile >>"coverage/cover.out"
	rm *.coverprofile
	go tool cover -html=coverage/cover.out -o=coverage/cover.html

.PHONY: install
install: build docs
	install -d -m 0755 $(DESTDIR)$(BINDIR)
	install -d -m 0755 $(DESTDIR)$(WWCHROOTDIR)
	install -d -m 0755 $(DESTDIR)$(WWPROVISIONDIR)
	install -d -m 0755 $(DESTDIR)$(WWOVERLAYDIR)/wwinit/rootfs/$(WWCLIENTDIR)
	install -d -m 0755 $(DESTDIR)$(WWOVERLAYDIR)/host/rootfs/$(TFTPDIR)/warewulf/
	install -d -m 0755 $(DESTDIR)$(WWCONFIGDIR)/examples
	install -d -m 0755 $(DESTDIR)$(WWCONFIGDIR)/ipxe
	install -d -m 0755 $(DESTDIR)$(WWCONFIGDIR)/grub
	install -d -m 0755 $(DESTDIR)$(WWCONFIGDIR)/dracut/modules.d/90wwinit
	install -d -m 0755 $(DESTDIR)$(BASHCOMPDIR)
	install -d -m 0755 $(DESTDIR)$(MANDIR)/man1
	install -d -m 0755 $(DESTDIR)$(MANDIR)/man5
	install -d -m 0755 $(DESTDIR)$(WWDOCDIR)
	install -d -m 0755 $(DESTDIR)$(FIREWALLDDIR)
	install -d -m 0755 $(DESTDIR)$(SYSTEMDDIR)
	install -d -m 0755 $(DESTDIR)$(IPXESOURCE)
	install -d -m 0755 $(DESTDIR)$(DATADIR)/warewulf
	# install dracut build script
	for f in etc/dracut/modules.d/90wwinit/*.sh; do install -m 0755 $$f $(DESTDIR)$(WWCONFIGDIR)/dracut/modules.d/90wwinit/; done
	# wwctl genconfig to get the compiled in paths to warewulf.conf
	test -f $(DESTDIR)$(WWCONFIGDIR)/warewulf.conf || ./wwctl --warewulfconf etc/warewulf.conf genconfig warewulfconf print> $(DESTDIR)$(WWCONFIGDIR)/warewulf.conf
	test -f $(DESTDIR)$(WWCONFIGDIR)/nodes.conf || install -m 0644 etc/nodes.conf $(DESTDIR)$(WWCONFIGDIR)
	test -f $(DESTDIR)$(WWCONFIGDIR)/wwapic.conf || install -m 0644 etc/wwapic.conf $(DESTDIR)$(WWCONFIGDIR)
	test -f $(DESTDIR)$(WWCONFIGDIR)/wwapid.conf || install -m 0644 etc/wwapid.conf $(DESTDIR)$(WWCONFIGDIR)
	test -f $(DESTDIR)$(WWCONFIGDIR)/wwapird.conf || install -m 0644 etc/wwapird.conf $(DESTDIR)$(WWCONFIGDIR)
	test -f $(DESTDIR)$(DATADIR)/warewulf/defaults.conf || install -m 0644 etc/defaults.conf $(DESTDIR)$(DATADIR)/warewulf/defaults.conf
	for f in etc/examples/*.ww; do install -m 0644 $$f $(DESTDIR)$(WWCONFIGDIR)/examples/; done
	for f in etc/ipxe/*.ipxe; do install -m 0644 $$f $(DESTDIR)$(WWCONFIGDIR)/ipxe/; done
	install -m 0644 etc/grub/grub.cfg.ww $(DESTDIR)$(WWCONFIGDIR)/grub/grub.cfg.ww
	install -m 0644 etc/grub/chainload.ww $(DESTDIR)$(WWOVERLAYDIR)/host/rootfs$(TFTPDIR)/warewulf/grub.cfg.ww
	(cd overlays && find * -type f -exec install -D -m 0644 {} $(DESTDIR)$(WWOVERLAYDIR)/{} \;)
	(cd overlays && find * -type d -exec mkdir -pv $(DESTDIR)$(WWOVERLAYDIR)/{} \;)
	(cd overlays && find * -type l -exec cp -av {} $(DESTDIR)$(WWOVERLAYDIR)/{} \;)
	chmod 0755 $(DESTDIR)$(WWOVERLAYDIR)/wwinit/rootfs/init
	chmod 0755 $(DESTDIR)$(WWOVERLAYDIR)/wwinit/rootfs/$(WWCLIENTDIR)/wwinit
	chmod 0600 $(DESTDIR)$(WWOVERLAYDIR)/wwinit/rootfs/etc/ssh/ssh*
	chmod 0600 $(DESTDIR)$(WWOVERLAYDIR)/wwinit/rootfs/etc/NetworkManager/system-connections/ww4-managed.ww
	chmod 0644 $(DESTDIR)$(WWOVERLAYDIR)/wwinit/rootfs/etc/ssh/ssh*.pub.ww
	chmod 0600 $(DESTDIR)$(WWOVERLAYDIR)/wwinit/rootfs/warewulf/config.ww
	chmod 0750 $(DESTDIR)$(WWOVERLAYDIR)/host/rootfs
	install -m 0755 wwctl $(DESTDIR)$(BINDIR)
	install -m 0755 wwclient $(DESTDIR)$(WWOVERLAYDIR)/wwinit/rootfs/$(WWCLIENTDIR)/wwclient
	install -m 0755 wwapic $(DESTDIR)$(BINDIR)
	install -m 0755 wwapid $(DESTDIR)$(BINDIR)
	install -m 0755 wwapird $(DESTDIR)$(BINDIR)
	install -m 0644 include/firewalld/warewulf.xml $(DESTDIR)$(FIREWALLDDIR)
	install -m 0644 include/systemd/warewulfd.service $(DESTDIR)$(SYSTEMDDIR)
	install -m 0644 LICENSE.md $(DESTDIR)$(WWDOCDIR)
	install -m 0644 etc/bash_completion.d/wwctl $(DESTDIR)$(BASHCOMPDIR)/wwctl
	for f in docs/man/man1/*.1.gz; do install -m 0644 $$f $(DESTDIR)$(MANDIR)/man1/; done
	for f in docs/man/man5/*.5.gz; do install -m 0644 $$f $(DESTDIR)$(MANDIR)/man5/; done

.PHONY: init
init:
	systemctl daemon-reload
	cp -r tftpboot/* $(WWTFTPDIR)/ipxe/
	restorecon -r $(WWTFTPDIR)

.PHONY: dist
dist:
	rm -rf .dist/ $(WAREWULF)-$(VERSION).tar.gz
	mkdir -p .dist/$(WAREWULF)-$(VERSION)
	rsync -a --exclude=".github"  --exclude=".vscode" --exclude "*~" * .dist/$(WAREWULF)-$(VERSION)/
	scripts/get-version.sh >.dist/$(WAREWULF)-$(VERSION)/VERSION
	cd .dist; tar -czf ../$(WAREWULF)-$(VERSION).tar.gz $(WAREWULF)-$(VERSION)
	rm -rf .dist

.PHONY: reference
reference: wwctl
	mkdir -p userdocs/reference
	./wwctl --emptyconf genconfig reference userdocs/reference/

latexpdf: reference
	make -C userdocs latexpdf

.PHONY: cleanconfig
cleanconfig:
	rm -f $(config)
	rm -rf etc/bash_completion.d/
	rm -rf etc/defaults.conf

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
	rm -f wwapi{c,d,rd}
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
clean: cleanconfig cleantest cleandist cleantools cleanmake cleanbin cleandocs

ifndef OFFLINE_BUILD
wwctl: vendor
wwclient: vendor
update_configuration: vendor
wwapid: vendor
wwapic: vendor
wwapird: vendor
dist: vendor

lint: $(GOLANGCI_LINT)

protofiles = internal/pkg/api/routes/wwapiv1/routes.pb.go \
	internal/pkg/api/routes/wwapiv1/routes.pb.gw.go \
	internal/pkg/api/routes/wwapiv1/routes_grpc.pb.go
.PHONY: proto
proto: $(protofiles)

routes_proto = internal/pkg/api/routes/v1/routes.proto
$(protofiles): $(routes_proto) $(PROTOC) $(PROTOC_GEN_GRPC_GATEWAY) $(PROTOC_GEN_GO) $(PROTOC_GEN_GO_GRPC)
	PATH=$(TOOLS_BIN):$(PATH) $(PROTOC) \
		-I /usr/include -I $(shell dirname $(routes_proto)) -I=. \
		--grpc-gateway_opt logtostderr=true \
		--go_out=. \
		--go-grpc_out=. \
		--grpc-gateway_out=. \
		routes.proto

.PHONY: cleanproto
cleanproto:
	rm -f $(protofiles)

clean: cleanvendor
endif
