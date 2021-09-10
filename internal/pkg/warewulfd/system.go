package warewulfd

import (
	"net/http"

	"github.com/hpcng/warewulf/internal/pkg/config"
)

func SystemOverlaySend(w http.ResponseWriter, req *http.Request) {
	node, err := getSanity(req)
	if err != nil {
		w.WriteHeader(404)
		daemonLogf("ERROR: %s\n", err)
		return
	}

	if node.SystemOverlay.Defined() {
		fileName := config.SystemOverlayImage(node.Id.Get())

		err := sendFile(w, fileName, node.Id.Get())
		if err != nil {
			daemonLogf("ERROR: %s\n", err)
		}
	} else {
		w.WriteHeader(503)
		daemonLogf("WARNING: No 'system system-overlay' set for node %s\n", node.Id.Get())
	}
}
