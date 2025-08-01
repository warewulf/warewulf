package warewulfd

import (
	"net/http"
	"net/netip"
	"strings"

	"github.com/pkg/errors"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

type parserInfo struct {
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

func parseReq(req *http.Request) (parserInfo, error) {
	var ret parserInfo

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

		if stage == "ipxe" || stage == "provision" {
			ret.stage = "ipxe"
		} else if stage == "kernel" {
			ret.stage = "kernel"
		} else if stage == "image" {
			ret.stage = "image"
		} else if stage == "overlay-system" {
			ret.stage = "system"
		} else if stage == "overlay-runtime" {
			ret.stage = "runtime"
		} else if stage == "efiboot" {
			ret.stage = "efiboot"
		} else if stage == "initramfs" {
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
