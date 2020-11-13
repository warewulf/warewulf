package warewulfd_responses

import (
	"github.com/hpcng/warewulf/internal/pkg/config"
	"log"
	"net/http"
)

func SystemOverlaySend(w http.ResponseWriter, req *http.Request) {
	config := config.New()

	node, err := getSanity(req)
	if err != nil {
		w.WriteHeader(404)
		log.Println(err)
		return
	}

	if node.SystemOverlay != "" {
		fileName := config.SystemOverlayImage(node.Fqdn)

		err := sendFile(w, fileName, node.Fqdn)
		if err != nil {
			log.Printf("ERROR: %s\n", err)
		} else {
			log.Printf("SEND:  %15s: %s\n", node.Fqdn, fileName)
		}
	} else {
		w.WriteHeader(503)
		log.Printf("ERROR: No 'system system-overlay' set for node %s\n", node.Fqdn)
	}

	return
}
