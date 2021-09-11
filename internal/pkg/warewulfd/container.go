package warewulfd

import (
	"net/http"

	"github.com/hpcng/warewulf/internal/pkg/container"
)

func ContainerSend(w http.ResponseWriter, req *http.Request) {
	node, err := getSanity(req)
	if err != nil {
		w.WriteHeader(404)
		daemonLogf("ERROR: %s\n", err)
		return
	}

	if node.ContainerName.Defined() {
		containerImage := container.ImageFile(node.ContainerName.Get())

		err = sendFile(w, containerImage, node.Id.Get())
		if err != nil {
			daemonLogf("ERROR: %s\n", err)
			w.WriteHeader(503)
		}
	} else {
		w.WriteHeader(503)
		daemonLogf("WARNING: No Container set for node %s\n", node.Id.Get())
	}
}
