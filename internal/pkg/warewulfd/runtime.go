package warewulfd

import (
	"net/http"
	"strings"

	nodepkg "github.com/hpcng/warewulf/internal/pkg/node"
	"github.com/hpcng/warewulf/internal/pkg/overlay"
	"github.com/hpcng/warewulf/internal/pkg/util"
	"github.com/hpcng/warewulf/internal/pkg/warewulfconf"
)

func RuntimeOverlaySend(w http.ResponseWriter, req *http.Request) {
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
		updateStatus(node.Id.Get(), "RUNTIME_OVERLAY", "BAD_ASSET", rinfo.ipaddr)
		return
	}

	conf, err := warewulfconf.New()
	if err != nil {
		daemonLogf("ERROR: Could not read Warewulf configuration file: %s\n", err)
		w.WriteHeader(503)
		return
	}

	if conf.Warewulf.Secure {
		if rinfo.remoteport >= 1024 {
			daemonLogf("DENIED: Connection coming from non-privledged port: %s\n", req.RemoteAddr)
			w.WriteHeader(401)
			return
		}
	}

	if len(node.RuntimeOverlay.GetSlice()) != 0 {
		fileName := overlay.OverlayImage(node.Id.Get(), node.RuntimeOverlay.GetSlice())

		updateStatus(node.Id.Get(), "RUNTIME_OVERLAY", node.RuntimeOverlay.Get(), strings.Split(req.RemoteAddr, ":")[0])

		if conf.Warewulf.AutobuildOverlays {
			oneoverlaynewer := false
			for _, overlayname := range node.RuntimeOverlay.GetSlice() {
				oneoverlaynewer = oneoverlaynewer || util.PathIsNewer(fileName, overlay.OverlaySourceDir(overlayname))
			}
			if !util.IsFile(fileName) || util.PathIsNewer(fileName, nodepkg.ConfigFile) || oneoverlaynewer {
				daemonLogf("BUILD: %15s: Runtime Overlay\n", node.Id.Get())
				_ = overlay.BuildOverlay(node, node.RuntimeOverlay.GetSlice())
			}
		}

		err := sendFile(w, fileName, node.Id.Get())
		if err != nil {
			daemonLogf("ERROR: %s\n", err)
		}

	} else {
		w.WriteHeader(503)
		daemonLogf("WARNING: No 'runtime overlay' set for node %s\n", node.Id.Get())
	}
}
