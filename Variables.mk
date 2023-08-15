-include Defaults.mk

# Linux distro (try and set to /etc/os-release ID)
OS_REL := $(shell sed -n "s/^ID\s*=\s*['"\""]\(.*\)['"\""]/\1/p" /etc/os-release)
OS ?= $(OS_REL)

# List of variables to save and replace in files
VARLIST := OS

# Project Information
VARLIST += WAREWULF VERSION RELEASE
WAREWULF ?= warewulf
VERSION ?= 4.5.x
GIT_TAG := $(shell test -e .git && git log -1 --format="%h")

ifdef GIT_TAG
  ifdef $(filter $(OS),ubuntu debian)
    RELEASE ?= 1.git_$(subst -,+,$(GIT_TAG))
  else
    RELEASE ?= 1.git_$(subst -,_,$(GIT_TAG))
  endif
else
  RELEASE ?= 1
endif

# Use LSB-compliant paths if OS is known
ifneq ($(OS),)
  USE_LSB_PATHS := true
endif

# Always default to GNU autotools default paths if PREFIX has been redefined
ifdef PREFIX
  USE_LSB_PATHS := false
endif

# System directory paths
VARLIST += PREFIX BINDIR SYSCONFDIR SRVDIR DATADIR MANDIR DOCDIR LOCALSTATEDIR
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

# OS-Specific Service Locations
VARLIST += TFTPDIR FIREWALLDDIR SYSTEMDDIR BASHCOMPDIR
SYSTEMDDIR ?= /usr/lib/systemd/system
BASHCOMPDIR ?= /etc/bash_completion.d
FIREWALLDDIR ?= /usr/lib/firewalld/services
ifeq ($(OS),suse)
  TFTPDIR ?= /srv/tftpboot
endif
ifeq ($(OS),ubuntu)
  TFTPDIR ?= /srv/tftp
endif
# Default to Red Hat / Rocky Linux
TFTPDIR ?= /var/lib/tftpboot

# Warewulf directory paths
VARLIST += WWCLIENTDIR WWCONFIGDIR WWPROVISIONDIR WWOVERLAYDIR WWCHROOTDIR WWTFTPDIR WWDOCDIR WWDATADIR
WWCONFIGDIR := $(SYSCONFDIR)/$(WAREWULF)
WWPROVISIONDIR := $(SRVDIR)/$(WAREWULF)
WWOVERLAYDIR := $(LOCALSTATEDIR)/$(WAREWULF)/overlays
WWCHROOTDIR := $(LOCALSTATEDIR)/$(WAREWULF)/chroots
WWTFTPDIR := $(TFTPDIR)/$(WAREWULF)
WWDOCDIR := $(DOCDIR)/$(WAREWULF)
WWDATADIR := $(DATADIR)/$(WAREWULF)
WWCLIENTDIR ?= /warewulf

# auto installed tooling
TOOLS_DIR := .tools
TOOLS_BIN := $(TOOLS_DIR)/bin
CONFIG := $(shell pwd)

# tools
GO_TOOLS_BIN := $(addprefix $(TOOLS_BIN)/, $(notdir $(GO_TOOLS)))
GO_TOOLS_VENDOR := $(addprefix vendor/, $(GO_TOOLS))
GOLANGCI_LINT := $(TOOLS_BIN)/golangci-lint
GOLANGCI_LINT_VERSION := v1.53.2
PROTOC_GEN_GO := $(TOOLS_BIN)/protoc-gen-go
PROTOC_GEN_GO_GRPC := $(TOOLS_BIN)/protoc-gen-go-grpc
PROTOC := $(TOOLS_BIN)/protoc
PROTOC_URL := https://github.com/protocolbuffers/protobuf/releases/download/v24.0/protoc-24.0-linux-x86_64.zip
#PROTOC_URL := https://github.com/protocolbuffers/protobuf/releases/download/v24.0/protoc-24.0-linux-aarch_64.zip
PROTOC_GEN_GRPC_GATEWAY := $(TOOLS_BIN)/protoc-gen-grpc-gateway
PROTOC_GEN_GRPC_GATEWAY_URL := https://github.com/grpc-ecosystem/grpc-gateway/releases/download/v2.16.2/protoc-gen-grpc-gateway-v2.16.2-linux-x86_64
#PROTOC_GEN_GRPC_GATEWAY_URL := https://github.com/grpc-ecosystem/grpc-gateway/releases/download/v2.16.2/protoc-gen-grpc-gateway-v2.16.2-linux-arm64

# helper functions
godeps=$(shell go list -deps -f '{{if not .Standard}}{{ $$dep := . }}{{range .GoFiles}}{{$$dep.Dir}}/{{.}} {{end}}{{end}}' $(1) | sed "s%${PWD}/%%g")

# use GOPROXY for older git clients and speed up downloads
GOPROXY ?= https://proxy.golang.org
export GOPROXY

# built tags needed for wwbuild binary
WW_GO_BUILD_TAGS := containers_image_openpgp containers_image_ostree

Defaults.mk:
	printf " $(foreach V,$(VARLIST),$V := $(strip $($V))\n)" >Defaults.mk
