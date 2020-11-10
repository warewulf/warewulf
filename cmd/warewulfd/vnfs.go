package main

import (
	"github.com/hpcng/warewulf/internal/pkg/vnfs"
	"log"
	"net/http"
)

func vnfsSend(w http.ResponseWriter, req *http.Request) {

	node, err := getSanity(req)
	if err != nil {
		w.WriteHeader(404)
		log.Panicln(err)
		return
	}

	if node.Vnfs != "" {
		v := vnfs.New(node.Vnfs)

		err := sendFile(w, v.Image(), node.Fqdn)
		if err != nil {
			log.Printf("ERROR: %s\n", err)
		} else {
			log.Printf("SEND:  %15s: %s\n", node.Fqdn, v.Image())
		}
	} else {
		w.WriteHeader(503)
		log.Printf("ERROR: No VNFS set for node %s\n", node.Fqdn)
	}

	return
}
