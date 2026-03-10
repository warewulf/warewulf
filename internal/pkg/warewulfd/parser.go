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
	conf       *warewulfconf.WarewulfYaml
	rinfo      parsedRequest
	remoteNode node.Node
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

	remoteNode, err := GetNodeOrSetDiscoverable(rinfo.hwaddr, conf.Warewulf.AutobuildOverlays())
	if err != nil && err != node.ErrNoUnconfigured {
		wwlog.ErrorExc(err, "")
		w.WriteHeader(http.StatusServiceUnavailable)
		return nil, err
	}

	if remoteNode.AssetKey != "" && remoteNode.AssetKey != rinfo.assetkey {
		w.WriteHeader(http.StatusUnauthorized)
		wwlog.Denied("incorrect asset key: node %s: %s", remoteNode.Id(), rinfo.assetkey)
		updateStatus(remoteNode.Id(), rinfo.stage, "BAD_ASSET", rinfo.ipaddr)
		return nil, fmt.Errorf("incorrect asset key")
	}

	return &requestContext{
		conf:       conf,
		rinfo:      rinfo,
		remoteNode: remoteNode,
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

func parseHwaddr(hwaddr string) string {
	hwaddr = strings.ReplaceAll(hwaddr, "-", ":")
	hwaddr = strings.ToLower(hwaddr)
	return hwaddr
}

// parseRequest extracts provisioning parameters from an HTTP request. The
// stage and hwaddr are taken from the URL path, with fallbacks to query
// parameters and ARP cache lookup for hwaddr. Stage aliases (e.g.
// "container", "overlay-system") are normalized to their canonical names.
//
// See userdocs/server/routes.rst for more information.
func parseRequest(req *http.Request) (parsedRequest, error) {
	var ret parsedRequest

	path_parts := strings.Split(req.URL.Path, "/")

	// initial stage passed in the url path /[stage]/hwaddr
	if len(path_parts) < 2 {
		return ret, errors.New("path missing initial stage: " + req.URL.Path)
	}
	ret.stage = path_parts[1]

	// prefer stage from query string for provision stage
	if ret.stage == "provision" && len(req.URL.Query()["stage"]) > 0 {
		ret.stage = req.URL.Query()["stage"][0]
	}

	// map the requested stage to a known stage
	switch ret.stage {
	case "provision":
		ret.stage = "ipxe"
	case "container":
		ret.stage = "image"
	case "overlay-system":
		ret.stage = "system"
	case "overlay-runtime":
		ret.stage = "runtime"
	}

	if ret.stage == "" {
		return ret, errors.New("no stage specified: " + req.URL.RawQuery)
	}

	if len(path_parts) > 2 {
		if ret.stage == "efiboot" {
			// /efiboot/{file}: no wwid in path; identified via ARP
			ret.efifile = strings.Join(path_parts[2:], "/")
		} else {
			ret.hwaddr = parseHwaddr(path_parts[2])
		}
	}

	remoteAddrPort, err := netip.ParseAddrPort(req.RemoteAddr)
	if err != nil {
		return ret, errors.New("could not parse remote address")
	}
	ret.ipaddr = remoteAddrPort.Addr().String()
	if ret.ipaddr == "" {
		return ret, errors.New("could not obtain ipaddr from HTTP request")
	}
	ret.remoteport = int(remoteAddrPort.Port())
	if ret.remoteport == 0 {
		return ret, errors.New("could not obtain remote port from HTTP request: " + req.RemoteAddr)
	}

	if ret.hwaddr == "" && len(req.URL.Query()["wwid"]) > 0 {
		ret.hwaddr = parseHwaddr(req.URL.Query()["wwid"][0])
	}
	if ret.hwaddr == "" {
		if hwaddr := parseHwaddr(ArpFind(ret.ipaddr)); hwaddr != "" {
			ret.hwaddr = hwaddr
			wwlog.Verbose("using %s from arp cache for %s", ret.hwaddr, ret.ipaddr)
		}
	}
	if ret.hwaddr == "" {
		return ret, errors.New("unable to determine wwid: " + req.URL.RawQuery)
	}

	if len(req.URL.Query()["assetkey"]) > 0 {
		ret.assetkey = req.URL.Query()["assetkey"][0]
	}
	if len(req.URL.Query()["uuid"]) > 0 {
		ret.uuid = req.URL.Query()["uuid"][0]
	}
	if len(req.URL.Query()["overlay"]) > 0 {
		ret.overlay = req.URL.Query()["overlay"][0]
	}
	if len(req.URL.Query()["compress"]) > 0 {
		ret.compress = req.URL.Query()["compress"][0]
	}
	if ret.efifile == "" && len(req.URL.Query()["file"]) > 0 {
		ret.efifile = req.URL.Query()["file"][0]
	}

	return ret, nil
}
