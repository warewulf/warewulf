-include Defaults.mk

# Linux distro (try and set to /etc/os-release ID)
OS_REL := $(shell sh -c '. /etc/os-release; echo $ID')
OS ?= $(OS_REL)

ARCH_REL := $(shell uname -p)
ARCH ?= $(ARCH_REL)

# List of variables to save and replace in files
VARLIST := OS

# Project Information
VARLIST += WAREWULF VERSION RELEASE
WAREWULF ?= warewulf
VERSION ?= $(shell scripts/rpm-version.sh $$(scripts/get-version.sh))
RELEASE ?= $(shell scripts/rpm-release.sh $$(scripts/get-version.sh))

# Use LSB-compliant paths if OS is known
ifneq ($(OS),)
  USE_LSB_PATHS := true
endif

# Always default to GNU autotools default paths if PREFIX has been redefined
ifdef PREFIX
  USE_LSB_PATHS := false
endif

# System directory paths
VARLIST += PREFIX BINDIR SYSCONFDIR SRVDIR DATADIR MANDIR DOCDIR LOCALSTATEDIR RELEASE CACHEDIR
PREFIX ?= /usr/local
BINDIR ?= $(PREFIX)/bin
SYSCONFDIR ?= $(PREFIX)/etc
DATADIR ?= $(PREFIX)/share
MANDIR ?= $(DATADIR)/man
DOCDIR ?= $(DATADIR)/doc

ifeq ($(USE_LSB_PATHS),true)
  SRVDIR ?= /srv
  LOCALSTATEDIR ?= /var/local
else
  SRVDIR ?= $(PREFIX)/srv
  LOCALSTATEDIR ?= $(PREFIX)/var
endif
CACHEDIR ?= $(LOCALSTATEDIR)/cache

# OS-Specific Service Locations
VARLIST += TFTPDIR FIREWALLDDIR SYSTEMDDIR BASHCOMPDIR LOGROTATEDIR DRACUTMODDIR
SYSTEMDDIR ?= /usr/lib/systemd/system
BASHCOMPDIR ?= /etc/bash_completion.d
FIREWALLDDIR ?= /usr/lib/firewalld/services
LOGROTATEDIR ?= /etc/logrotate.d
DRACUTMODDIR ?= /usr/lib/dracut/modules.d
SOSPLUGINS ?= /usr/lib/python3.9/site-packages/sos/report/plugins
ifeq ($(OS),suse)
  TFTPDIR ?= /srv/tftpboot
endif
ifeq ($(OS),ubuntu)
  TFTPDIR ?= /srv/tftp
endif
# Default to Red Hat / Rocky Linux
TFTPDIR ?= /var/lib/tftpboot

# Warewulf directory paths
VARLIST += WWCLIENTDIR WWCONFIGDIR WWPROVISIONDIR WWOVERLAYDIR WWCHROOTDIR WWTFTPDIR WWDOCDIR IPXESOURCE SOSPLUGINS
WWCONFIGDIR ?= $(SYSCONFDIR)/$(WAREWULF)
WWPROVISIONDIR ?= $(LOCALSTATEDIR)/$(WAREWULF)/provision
WWOVERLAYDIR ?= $(LOCALSTATEDIR)/$(WAREWULF)/overlays
WWCHROOTDIR ?= $(LOCALSTATEDIR)/$(WAREWULF)/chroots
WWTFTPDIR ?= $(TFTPDIR)/$(WAREWULF)
WWDOCDIR ?= $(DOCDIR)/$(WAREWULF)
WWCLIENTDIR ?= /warewulf

CONFIG := $(shell pwd)

IPXESOURCE ?= $(PREFIX)/share/ipxe

# helper functions
godeps=$(shell 2>/dev/null go list -mod vendor -deps -f '{{if not .Standard}}{{ $$dep := . }}{{range .GoFiles}}{{$$dep.Dir}}/{{.}} {{end}}{{end}}' $(1) | sed "s%$(shell pwd)/%%g")

# use GOPROXY for older git clients and speed up downloads
GOPROXY ?= https://proxy.golang.org
export GOPROXY

# built tags needed for wwbuild binary
WW_GO_BUILD_TAGS := containers_image_openpgp containers_image_ostree

.PHONY: defaults
defaults: Defaults.mk

Defaults.mk:
	printf " $(foreach V,$(VARLIST),$V := $(strip $($V))\n)" >Defaults.mk
