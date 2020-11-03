package main

import (
	"fmt"
	"log"
	"net/http"
)

func overlay(w http.ResponseWriter, req *http.Request) {

	node, err := getSanity(req)
	if err != nil {
		w.WriteHeader(404)
		log.Println(err)
		return
	}

	if node.Overlay!= "" {
		fileName := fmt.Sprintf("%s/provision/overlays/%s.img", LocalStateDir, node.Fqdn)

		err := sendFile(w, fileName, node.Fqdn)
		if err != nil {
			log.Println(err)
		}
	} else {
		w.WriteHeader(503)
		log.Printf("No Overlay set for node %s\n", node.Fqdn)
	}

	return
}