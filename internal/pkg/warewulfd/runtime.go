package warewulfd

import (
	"fmt"
	"github.com/hpcng/warewulf/internal/pkg/config"
	"github.com/hpcng/warewulf/internal/pkg/node"
	"github.com/hpcng/warewulf/internal/pkg/warewulfconf"
	"log"
	"net/http"
	"strconv"
	"strings"
)

func RuntimeOverlaySend(w http.ResponseWriter, req *http.Request) {
	conf, err := warewulfconf.New()
	if err != nil {
		log.Printf("Could not read Warewulf configuration file: %s\n", err)
		w.WriteHeader(503)
		return
	}

	nodes, err := node.New()
	if err != nil {
		log.Printf("Could not read node configuration file: %s\n", err)
		w.WriteHeader(503)
		return
	}

	remote := strings.Split(req.RemoteAddr, ":")
	port, err := strconv.Atoi(remote[1])
	if err != nil {
		log.Printf("Could not convert port to integer: %s\n", remote[1])
		w.WriteHeader(503)
		return
	}

	if err != nil {
		fmt.Printf("ERROR: Could not load configuration file: %s\n", err)
		return
	}

	if conf.Warewulf.Secure == true {
		if port >= 1024 {
			log.Panicf("DENIED: Connection coming from non-privledged port: %s\n", req.RemoteAddr)
			w.WriteHeader(401)
			return
		}
	}

	node, err := nodes.FindByIpaddr(remote[0])
	if err != nil {
		fmt.Printf("Could not find node by IP address: %s\n", remote[0])
		w.WriteHeader(404)
		return
	}

	if node.Id.Defined() == false {
		log.Printf("UNKNOWN: %15s: %s\n", remote[0], req.URL.Path)
		w.WriteHeader(404)
		return
	} else {
		log.Printf("REQ:   %15s: %s\n", node.Id.Get(), req.URL.Path)
	}

	if node.RuntimeOverlay.Defined() == true {
		fileName := config.RuntimeOverlayImage(node.Id.Get())

		err := sendFile(w, fileName, node.Id.Get())
		if err != nil {
			log.Printf("ERROR: %s\n", err)
		} else {
			log.Printf("SEND:  %15s: %s\n", node.Id.Get(), fileName)
		}
	} else {
		w.WriteHeader(503)
		log.Printf("ERROR: No 'runtime system-overlay' set for node %s\n", node.Id.Get())
	}

	return
}
