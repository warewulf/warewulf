package warewulfd

import (
	"net/http"

	"github.com/hpcng/warewulf/internal/pkg/config"
	"github.com/hpcng/warewulf/internal/pkg/node"
	"github.com/hpcng/warewulf/internal/pkg/overlay"
	"github.com/hpcng/warewulf/internal/pkg/util"
)

func SystemOverlaySend(w http.ResponseWriter, req *http.Request) {
	n, err := getSanity(req)
	if err != nil {
		w.WriteHeader(404)
		daemonLogf("ERROR: %s\n", err)
		return
	}

	if n.SystemOverlay.Defined() {
		fileName := config.SystemOverlayImage(n.Id.Get())

		if !util.IsFile(fileName) || util.PathIsNewer(fileName, node.ConfigFile) {
			daemonLogf("BUILD: %15s: System Overlay\n", n.Id.Get())
			_ = overlay.BuildSystemOverlay([]node.NodeInfo{n})
		}

		err := sendFile(w, fileName, n.Id.Get())
		if err != nil {
			daemonLogf("ERROR: %s\n", err)
		}
	} else {
		w.WriteHeader(503)
		daemonLogf("WARNING: No 'system system-overlay' set for node %s\n", n.Id.Get())
	}
}
