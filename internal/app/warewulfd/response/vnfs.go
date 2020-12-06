package response

import (
	"github.com/hpcng/warewulf/internal/pkg/vnfs"
	"log"
	"net/http"
)

func VnfsSend(w http.ResponseWriter, req *http.Request) {

	node, err := getSanity(req)
	if err != nil {
		w.WriteHeader(404)
		log.Panicln(err)
		return
	}

	if node.Vnfs.Defined() == true {
		vnfsImage := vnfs.ImageFile(node.Vnfs.Get())

		err = sendFile(w, vnfsImage, node.Id.Get())
		if err != nil {
			log.Printf("ERROR1: %s\n", err)
			w.WriteHeader(503)
		} else {
			log.Printf("SEND:  %15s: %s\n", node.Id.Get(), vnfsImage)
		}
	} else {
		w.WriteHeader(503)
		log.Printf("ERROR: No VNFS set for node %s\n", node.Id.Get())
	}

	return
}
