package warewulfd

import (
	"net/http"
	"strings"

	nodepkg "github.com/hpcng/warewulf/internal/pkg/node"
	"github.com/hpcng/warewulf/internal/pkg/overlay"
	"github.com/hpcng/warewulf/internal/pkg/util"
	"github.com/hpcng/warewulf/internal/pkg/warewulfconf"
)

func SystemOverlaySend(w http.ResponseWriter, req *http.Request) {
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
		updateStatus(node.Id.Get(), "SYSTEM_OVERLAY", "BAD_ASSET", rinfo.ipaddr)
		return
	}

	conf, err := warewulfconf.New()
	if err != nil {
		daemonLogf("ERROR: Could not read Warewulf configuration file: %s\n", err)
		w.WriteHeader(503)
		return
	}

	if node.SystemOverlay.Defined() {
		fileName := overlay.OverlayImage(node.Id.Get(), node.SystemOverlay.Get())

		updateStatus(node.Id.Get(), "SYSTEM_OVERLAY", node.SystemOverlay.Get()+".img", strings.Split(req.RemoteAddr, ":")[0])

		if conf.Warewulf.AutobuildOverlays {
			if !util.IsFile(fileName) || util.PathIsNewer(fileName, nodepkg.ConfigFile) || util.PathIsNewer(fileName, overlay.OverlaySourceDir(node.SystemOverlay.Get())) {
				daemonLogf("BUILD: %15s: System Overlay\n", node.Id.Get())
				_ = overlay.BuildOverlay(node, node.SystemOverlay.Get())
			}
		}

		err := sendFile(w, fileName, node.Id.Get())
		if err != nil {
			daemonLogf("ERROR: %s\n", err)
		}

	} else {
		w.WriteHeader(503)
		daemonLogf("WARNING: No 'system system-overlay' set for node %s\n", node.Id.Get())
	}
}
