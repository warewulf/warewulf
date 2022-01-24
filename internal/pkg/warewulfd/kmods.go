package warewulfd

import (
	"net/http"
	"strings"

	"github.com/hpcng/warewulf/internal/pkg/kernel"
)

func KmodsSend(w http.ResponseWriter, req *http.Request) {
	rinfo, err := parseReq(req)
	if err != nil {
		w.WriteHeader(404)
		daemonLogf("ERROR: %s\n", err)
		return
	}
	node, err := GetNode(rinfo.hwaddr)
	if err != nil {
		w.WriteHeader(403)
		daemonLogf("ERROR(%s): %s\n", rinfo.hwaddr, err)
		return
	}

	if node.AssetKey.Defined() && node.AssetKey.Get() != rinfo.assetkey {
		w.WriteHeader(404)
		daemonLogf("ERROR: Incorrect asset key for node: %s\n", node.Id.Get())
		updateStatus(node.Id.Get(), "KMODS_OVERLAY", "BAD_ASSET", rinfo.ipaddr)
		return
	}

	if node.KernelVersion.Defined() {
		fileName := kernel.KmodsImage(node.KernelVersion.Get())

		err := sendFile(w, fileName, node.Id.Get())
		if err != nil {
			daemonLogf("ERROR: %s\n", err)
		}

		updateStatus(node.Id.Get(), "KMODS_OVERLAY", node.KernelVersion.Get()+".img", strings.Split(req.RemoteAddr, ":")[0])

	} else {
		w.WriteHeader(503)
		daemonLogf("WARNING: No 'kernel version' set for node %s\n", node.Id.Get())
	}
}
