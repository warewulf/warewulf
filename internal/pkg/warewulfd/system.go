package warewulfd

import (
	"net/http"
	"strings"

	"github.com/hpcng/warewulf/internal/pkg/config"
	"github.com/hpcng/warewulf/internal/pkg/node"
	"github.com/hpcng/warewulf/internal/pkg/overlay"
	"github.com/hpcng/warewulf/internal/pkg/util"
	"github.com/hpcng/warewulf/internal/pkg/warewulfconf"
)

func SystemOverlaySend(w http.ResponseWriter, req *http.Request) {
	conf, err := warewulfconf.New()
	if err != nil {
		daemonLogf("ERROR: Could not read Warewulf configuration file: %s\n", err)
		w.WriteHeader(503)
		return
	}

	n, err := getSanity(req)
	if err != nil {
		w.WriteHeader(404)
		daemonLogf("ERROR: %s\n", err)
		return
	}

	if n.SystemOverlay.Defined() {
		fileName := config.SystemOverlayImage(n.Id.Get())

		if conf.Warewulf.AutobuildOverlays {
			if !util.IsFile(fileName) || util.PathIsNewer(fileName, node.ConfigFile) || util.PathIsNewer(fileName, config.SystemOverlaySource(n.SystemOverlay.Get())) {
				daemonLogf("BUILD: %15s: System Overlay\n", n.Id.Get())
				_ = overlay.BuildSystemOverlay([]node.NodeInfo{n})
			}
		}

		err := sendFile(w, fileName, n.Id.Get())
		if err != nil {
			daemonLogf("ERROR: %s\n", err)
		}

		updateStatus(n.Id.Get(), "SYSTEM_OVERLAY", n.SystemOverlay.Get()+".img", strings.Split(req.RemoteAddr, ":")[0])

	} else {
		w.WriteHeader(503)
		daemonLogf("WARNING: No 'system system-overlay' set for node %s\n", n.Id.Get())
	}
}
