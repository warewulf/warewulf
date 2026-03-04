package warewulfd

import (
	"net/http"

	"github.com/warewulf/warewulf/internal/pkg/node"
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
	TLS           bool
	Tags          map[string]string
	NetDevs       map[string]*node.NetDev
}

func HandleProvision(w http.ResponseWriter, req *http.Request) {
	// Parse just enough to determine the stage
	rinfo, err := parseRequest(req)
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
	case "efiboot":
		handler = HandleEfiBoot
	case "shim":
		handler = HandleShim
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
