package main

import (
	"fmt"
	"github.com/hpcng/warewulf/internal/pkg/assets"
	"log"
	"net/http"
	"strconv"
	"strings"
)

func runtimeOverlay(w http.ResponseWriter, req *http.Request) {

	remote := strings.Split(req.RemoteAddr, ":")
	port, err := strconv.Atoi(remote[1])
	if err != nil {
		log.Printf("Could not convert port to integer: %s\n", remote[1])
		w.WriteHeader(503)
		return
	}

	if port >= 1024 {
		log.Panicf("DENIED: Connection coming from non-privledged port: %s\n", req.RemoteAddr)
		w.WriteHeader(401)
		return
	}

	node, err := assets.FindByIpaddr(remote[0])
	if err != nil {
		fmt.Printf("Could not find node by IP address: %s\n", remote[0])
		w.WriteHeader(404)
		return
	}

	if node.Fqdn == "" {
		log.Printf("UNKNOWN: %15s: %s\n", remote[0], req.URL.Path)
		w.WriteHeader(404)
		return
	} else {
		log.Printf("REQ:   %15s: %s\n", node.Fqdn, req.URL.Path)
	}

	if node.RuntimeOverlay != "" {
		fileName := fmt.Sprintf("%s/provision/overlays/runtime/%s.img", LocalStateDir, node.Fqdn)

		err := sendFile(w, fileName, node.Fqdn)
		if err != nil {
			log.Printf("ERROR: %s\n", err)
		} else {
			log.Printf("SEND:  %15s: %s\n", node.Fqdn, fileName)
		}
	} else {
		w.WriteHeader(503)
		log.Printf("ERROR: No 'runtime overlay' set for node %s\n", node.Fqdn)
	}

	return
}
