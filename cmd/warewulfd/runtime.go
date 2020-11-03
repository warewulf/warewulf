package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
)

func runtime(w http.ResponseWriter, req *http.Request) {

	node, err := getSanity(req)
	if err != nil {
		w.WriteHeader(404)
		log.Panicln(err)
		return
	}

	remote := strings.Split(req.RemoteAddr, ":")
	port, err := strconv.Atoi(remote[1])
	if err != nil {
		w.WriteHeader(404)
		log.Printf("Could not convert port to integer: %s\n", remote[1])
		return
	}

	if port >= 1024 {
		log.Panicf("DENIED: Connection coming from non-privledged port: %s\n", req.RemoteAddr)
		return
	}

	if node.Overlay!= "" {
		fileName := fmt.Sprintf("%s/provision/runtime/%s.img", LocalStateDir, node.Fqdn)

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