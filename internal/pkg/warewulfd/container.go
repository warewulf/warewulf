package warewulfd

import (
	"github.com/hpcng/warewulf/internal/pkg/container"
	"log"
	"net/http"
)

func ContainerSend(w http.ResponseWriter, req *http.Request) {

	node, err := getSanity(req)
	if err != nil {
		w.WriteHeader(404)
		log.Panicln(err)
		return
	}

	if node.ContainerName.Defined() == true {
		containerImage := container.ImageFile(node.ContainerName.Get())

		err = sendFile(w, containerImage, node.Id.Get())
		if err != nil {
			log.Printf("ERROR1: %s\n", err)
			w.WriteHeader(503)
		} else {
			log.Printf("SEND:  %15s: %s\n", node.Id.Get(), containerImage)
		}
	} else {
		w.WriteHeader(503)
		log.Printf("ERROR: No Container set for node %s\n", node.Id.Get())
	}

	return
}
