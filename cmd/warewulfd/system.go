package main

import (
	"fmt"
	"log"
	"net/http"
)

func systemOverlay(w http.ResponseWriter, req *http.Request) {

	node, err := getSanity(req)
	if err != nil {
		w.WriteHeader(404)
		log.Println(err)
		return
	}

	if node.SystemOverlay != "" {
		fileName := fmt.Sprintf("%s/provision/overlays/system/%s.img", LocalStateDir, node.Fqdn)

		err := sendFile(w, fileName, node.Fqdn)
		if err != nil {
			log.Printf("ERROR: %s\n", err)
		} else {
			log.Printf("SEND:  %15s: %s\n", node.Fqdn, fileName)
		}
	} else {
		w.WriteHeader(503)
		log.Printf("ERROR: No 'system overlay' set for node %s\n", node.Fqdn)
	}

	return
}
