package warewulfd

import (
	"log"
	"net/http"

	"github.com/hpcng/warewulf/internal/pkg/config"
)

func SystemOverlaySend(w http.ResponseWriter, req *http.Request) {
	node, err := getSanity(req)
	if err != nil {
		w.WriteHeader(404)
		log.Println(err)
		return
	}

	if node.SystemOverlay.Defined() {
		fileName := config.SystemOverlayImage(node.Id.Get())

		err := sendFile(w, fileName, node.Id.Get())
		if err != nil {
			log.Printf("ERROR: %s\n", err)
		} else {
			log.Printf("SEND:  %15s: %s\n", node.Id.Get(), fileName)
		}
	} else {
		w.WriteHeader(503)
		log.Printf("ERROR: No 'system system-overlay' set for node %s\n", node.Id.Get())
	}
}
