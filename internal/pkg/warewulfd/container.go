package warewulfd

import (
	"net/http"
	"strings"

	"github.com/hpcng/warewulf/internal/pkg/container"
)

func ContainerSend(w http.ResponseWriter, req *http.Request) {
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
		updateStatus(node.Id.Get(), "CONTAINER", "BAD_ASSET", rinfo.ipaddr)
		return
	}

	if node.ContainerName.Defined() {
		containerImage := container.ImageFile(node.ContainerName.Get())

		err = sendFile(w, containerImage, node.Id.Get())
		if err != nil {
			daemonLogf("ERROR: %s\n", err)
			w.WriteHeader(503)
		}

		updateStatus(node.Id.Get(), "CONTAINER", node.ContainerName.Get()+".img", strings.Split(req.RemoteAddr, ":")[0])

	} else {
		w.WriteHeader(503)
		daemonLogf("WARNING: No Container set for node %s\n", node.Id.Get())
	}
}
