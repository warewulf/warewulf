package main

import (
	"fmt"
	"log"
	"net/http"
	"path"
)

func vnfs(w http.ResponseWriter, req *http.Request) {

	node, err := getSanity(req)
	if err != nil {
		w.WriteHeader(404)
		log.Panicln(err)
		return
	}

	if node.Vnfs != "" {
		fileName := fmt.Sprintf("%s/provision/vnfs/%s.img.gz", LocalStateDir, path.Base(node.Vnfs))

		err := sendFile(w, fileName, node.Fqdn)
		if err != nil {
			log.Printf("ERROR: %s\n", err)
		} else {
			log.Printf("SEND:  %15s: %s\n", node.Fqdn, fileName)
		}
	} else {
		w.WriteHeader(503)
		log.Printf("ERROR: No VNFS set for node %s\n", node.Fqdn)
	}

	return
}
