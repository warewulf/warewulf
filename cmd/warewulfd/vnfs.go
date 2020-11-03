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
		fileName := fmt.Sprintf("%s/provision/bases/%s.img.gz", LocalStateDir, path.Base(node.Vnfs))

		err := sendFile(w, fileName, node.Fqdn)
		if err != nil {
			log.Println(err)
		}
	} else {
		w.WriteHeader(503)
		log.Printf("No VNFS set for node %s\n", node.Fqdn)
	}

	return
}