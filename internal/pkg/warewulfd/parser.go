package warewulfd

import (
	"fmt"
	"net/http"
	"net/netip"
	"strings"

	"github.com/pkg/errors"
	warewulfconf "github.com/warewulf/warewulf/internal/pkg/config"
	"github.com/warewulf/warewulf/internal/pkg/node"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

// requestContext holds the validated results of the parsed request
type requestContext struct {
	conf        *warewulfconf.WarewulfYaml
	rinfo       parsedRequest
	remoteNode  node.Node
	statusStage string
}

// initHandleRequest performs common initial request parsing, security checks,
// node lookup, and asset key validation. On error, it writes the HTTP error
// response and returns a non-nil error so the caller can simply return.
func initHandleRequest(w http.ResponseWriter, req *http.Request) (*requestContext, error) {
	wwlog.Debug("Requested URL: %s", req.URL.String())
	conf := warewulfconf.Get()
	rinfo, err := parseRequest(req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		wwlog.ErrorExc(err, "Bad status")
		return nil, err
	}

	wwlog.Debug("stage: %s", rinfo.stage)

	wwlog.Info("request from hwaddr:%s ipaddr:%s | stage:%s", rinfo.hwaddr, req.RemoteAddr, rinfo.stage)

	if (rinfo.stage == "runtime" || len(rinfo.overlay) > 0) && conf.Warewulf.Secure() {
		if rinfo.remoteport >= 1024 {
			wwlog.Denied("Non-privileged port: %s", req.RemoteAddr)
			w.WriteHeader(http.StatusUnauthorized)
			return nil, fmt.Errorf("non-privileged port")
		}
	}

	status_stages := map[string]string{
		"efiboot":   "EFI",
		"ipxe":      "IPXE",
		"kernel":    "KERNEL",
		"system":    "SYSTEM_OVERLAY",
		"runtime":   "RUNTIME_OVERLAY",
		"initramfs": "INITRAMFS"}

	statusStage := status_stages[rinfo.stage]

	remoteNode, err := GetNodeOrSetDiscoverable(rinfo.hwaddr, conf.Warewulf.AutobuildOverlays())
	if err != nil && err != node.ErrNoUnconfigured {
		wwlog.ErrorExc(err, "")
		w.WriteHeader(http.StatusServiceUnavailable)
		return nil, err
	}

	if remoteNode.AssetKey != "" && remoteNode.AssetKey != rinfo.assetkey {
		w.WriteHeader(http.StatusUnauthorized)
		wwlog.Denied("incorrect asset key: node %s: %s", remoteNode.Id(), rinfo.assetkey)
		updateStatus(remoteNode.Id(), statusStage, "BAD_ASSET", rinfo.ipaddr)
		return nil, fmt.Errorf("incorrect asset key")
	}

	return &requestContext{
		conf:        conf,
		rinfo:       rinfo,
		remoteNode:  remoteNode,
		statusStage: statusStage,
	}, nil
}

type parsedRequest struct {
	hwaddr     string
	ipaddr     string
	remoteport int
	assetkey   string
	uuid       string
	stage      string
	overlay    string
	efifile    string
	compress   string
}

func parseRequest(req *http.Request) (parsedRequest, error) {
	var ret parsedRequest

	url := strings.Split(req.URL.Path, "?")[0]
	path_parts := strings.Split(url, "/")

	if len(path_parts) < 3 {
		return ret, errors.New("unknown path components in GET")
	}

	// handle when stage was passed in the url path /[stage]/hwaddr
	stage := path_parts[1]
	hwaddr := ""
	if stage != "efiboot" {
		hwaddr = path_parts[2]
		hwaddr = strings.ReplaceAll(hwaddr, "-", ":")
		hwaddr = strings.ToLower(hwaddr)
	} else if len(path_parts) > 3 {
		ret.efifile = strings.Join(path_parts[2:], "/")
	} else {
		ret.efifile = path_parts[2]
	}
	ret.hwaddr = hwaddr
	remoteAddrPort, err := netip.ParseAddrPort(req.RemoteAddr)
	if err != nil {
		return ret, errors.New("could not parse remote address")
	}
	ret.ipaddr = remoteAddrPort.Addr().String()
	ret.remoteport = int(remoteAddrPort.Port())
	if len(req.URL.Query()["assetkey"]) > 0 {
		ret.assetkey = req.URL.Query()["assetkey"][0]
	}

	if len(req.URL.Query()["uuid"]) > 0 {
		ret.uuid = req.URL.Query()["uuid"][0]
	}

	if len(req.URL.Query()["stage"]) > 0 {
		ret.stage = req.URL.Query()["stage"][0]
	} else {

		switch stage {
		case "ipxe", "provision":
			ret.stage = "ipxe"
		case "kernel":
			ret.stage = "kernel"
		case "image", "container":
			ret.stage = "image"
		case "overlay-system":
			ret.stage = "system"
		case "overlay-runtime":
			ret.stage = "runtime"
		case "efiboot":
			ret.stage = "efiboot"
		case "initramfs":
			ret.stage = "initramfs"
		}
	}

	if len(req.URL.Query()["overlay"]) > 0 {
		ret.overlay = req.URL.Query()["overlay"][0]
	}
	if len(req.URL.Query()["compress"]) > 0 {
		ret.compress = req.URL.Query()["compress"][0]
	}
	if ret.stage == "" {
		return ret, errors.New("no stage encoded in GET")
	}
	if ret.hwaddr == "" {
		ret.hwaddr = ArpFind(ret.ipaddr)
		wwlog.Verbose("node mac not encoded, arp cache got %s for %s", ret.hwaddr, ret.ipaddr)
		if ret.hwaddr == "" {
			return ret, errors.New("no hwaddr encoded in GET")
		}
	}
	if ret.ipaddr == "" {
		return ret, errors.New("could not obtain ipaddr from HTTP request")
	}
	if ret.remoteport == 0 {
		return ret, errors.New("could not obtain remote port from HTTP request: " + req.RemoteAddr)
	}

	return ret, nil
}
