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

		v, err := vnfs.Load(node.Vnfs.Get())
		if err != nil {
			w.WriteHeader(503)
			log.Printf("ERROR: Could not load VNFS: %s\n", node.Fqdn.Get())
			return
		}

		err = sendFile(w, v.Image, node.Fqdn.Get())
		if err != nil {
			log.Printf("ERROR1: %s\n", err)
		} else {
			log.Printf("SEND:  %15s: %s\n", node.Fqdn.Get(), v.Image)
		}
	} else {
		w.WriteHeader(503)
		log.Printf("ERROR: No VNFS set for node %s\n", node.Fqdn.Get())
	}

	return
}
