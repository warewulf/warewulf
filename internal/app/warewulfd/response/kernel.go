package response

import (
	"github.com/hpcng/warewulf/internal/pkg/kernel"
	"log"
	"net/http"
)

func KernelSend(w http.ResponseWriter, req *http.Request) {

	node, err := getSanity(req)
	if err != nil {
		w.WriteHeader(404)
		log.Println(err)
		return
	}

	if node.KernelVersion.Defined() == true {
		fileName := kernel.KernelImage(node.KernelVersion.Get())

		err := sendFile(w, fileName, node.Id.Get())
		if err != nil {
			log.Printf("ERROR: %s\n", err)
		} else {
			log.Printf("SEND:  %15s: %s\n", node.Id.Get(), fileName)
		}

	} else {
		w.WriteHeader(503)
		log.Printf("ERROR: No 'kernel version' set for node %s\n", node.Id.Get())
	}

	return
}
