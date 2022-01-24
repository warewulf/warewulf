package warewulfd

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

type parserInfo struct {
	hwaddr     string
	ipaddr     string
	remoteport int
	assetkey   string
	uuid       string
}

func parseReq(req *http.Request) (parserInfo, error) {
	var ret parserInfo

	url := strings.Split(req.URL.Path, "?")[0]
	hwaddr := strings.Split(url, "/")[2]
	hwaddr = strings.ReplaceAll(hwaddr, "-", ":")
	hwaddr = strings.ToUpper(hwaddr)

	ret.hwaddr = hwaddr
	ret.ipaddr = strings.Split(req.RemoteAddr, ":")[0]
	ret.remoteport, _ = strconv.Atoi(strings.Split(req.RemoteAddr, ":")[1])

	if len(req.URL.Query()["assetkey"]) > 0 {
		ret.assetkey = req.URL.Query()["assetkey"][0]
	}

	if len(req.URL.Query()["uuid"]) > 0 {
		ret.uuid = req.URL.Query()["uuid"][0]
	}

	if ret.hwaddr == "" {
		return ret, errors.New("no hwaddr encoded in GET")
	}
	if ret.ipaddr == "" {
		return ret, errors.New("could not obtain ipaddr from HTTP request")
	}
	if ret.remoteport == 0 {
		return ret, errors.New("could not obtain remote port from HTTP request: " + req.RemoteAddr)
	}

	return ret, nil
}
