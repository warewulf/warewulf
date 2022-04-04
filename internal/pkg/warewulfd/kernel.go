package warewulfd

import (
	"net/http"
	"path"
	"strings"

	"github.com/hpcng/warewulf/internal/pkg/container"
	"github.com/hpcng/warewulf/internal/pkg/kernel"
)

func KernelSend(w http.ResponseWriter, req *http.Request) {
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
		updateStatus(node.Id.Get(), "KERNEL", "BAD_ASSET", rinfo.ipaddr)
		return
	}

	var fileName string
	if node.Kernel.Override.Defined() {
		fileName = kernel.KernelImage(node.Kernel.Override.Get())

	} else if node.ContainerName.Defined() {
		fileName = container.KernelFind(node.ContainerName.Get())

	} else {
		w.WriteHeader(503)
		daemonLogf("WARNING: No 'kernel version' set for node %s\n", node.Id.Get())

		return
	}

	updateStatus(node.Id.Get(), "KERNEL", path.Base(fileName), strings.Split(req.RemoteAddr, ":")[0])

	err = sendFile(w, fileName, node.Id.Get())
	if err != nil {
		daemonLogf("ERROR: %s\n", err)
	}

}
