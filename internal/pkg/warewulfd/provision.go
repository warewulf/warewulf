package warewulfd

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"net/netip"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	warewulfconf "github.com/warewulf/warewulf/internal/pkg/config"
	"github.com/warewulf/warewulf/internal/pkg/image"
	"github.com/warewulf/warewulf/internal/pkg/kernel"
	"github.com/warewulf/warewulf/internal/pkg/node"
	"github.com/warewulf/warewulf/internal/pkg/overlay"
	"github.com/warewulf/warewulf/internal/pkg/util"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

type templateVars struct {
	Message       string
	WaitTime      string
	Hostname      string
	Fqdn          string
	Id            string
	Cluster       string
	ImageName     string
	Ipxe          string
	Hwaddr        string
	Ipaddr        string
	Ipaddr6       string
	Port          string
	Authority     string
	KernelArgs    string
	KernelVersion string
	Root          string
	Https         bool
	Tags          map[string]string
	NetDevs       map[string]*node.NetDev
}

func ProvisionSend(w http.ResponseWriter, req *http.Request) {
	wwlog.Debug("Requested URL: %s", req.URL.String())
	conf := warewulfconf.Get()
	rinfo, err := parseReq(req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		wwlog.ErrorExc(err, "Bad status")
		return
	}

	// Dispatch to the appropriate stage handler
	var handler http.HandlerFunc
	switch rinfo.stage {
	case "ipxe":
		handler = HandleIpxe
	case "kernel":
		handler = HandleKernel
	case "image":
		handler = HandleImage
	case "system":
		handler = HandleSystemOverlay
	case "runtime":
		handler = HandleRuntimeOverlay
	case "grub":
		handler = HandleGrub
	case "initramfs":
		handler = HandleInitramfs
	default:
		w.WriteHeader(http.StatusBadRequest)
		wwlog.Error("Unknown stage: %s", rinfo.stage)
		return
	}
	handler(w, req)
}
