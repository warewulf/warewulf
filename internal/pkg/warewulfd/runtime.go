package warewulfd

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/hpcng/warewulf/internal/pkg/config"
	"github.com/hpcng/warewulf/internal/pkg/node"
	"github.com/hpcng/warewulf/internal/pkg/warewulfconf"
)

func RuntimeOverlaySend(w http.ResponseWriter, req *http.Request) {
	conf, err := warewulfconf.New()
	if err != nil {
		daemonLogf("ERROR: Could not read Warewulf configuration file: %s\n", err)
		w.WriteHeader(503)
		return
	}

	nodes, err := node.New()
	if err != nil {
		daemonLogf("%s | ERROR: Could not read node configuration file: %s\n", err)
		w.WriteHeader(503)
		return
	}

	remote := strings.Split(req.RemoteAddr, ":")
	port, err := strconv.Atoi(remote[1])
	if err != nil {
		daemonLogf("%s | ERROR: Could not convert port to integer: %s\n", remote[1])
		w.WriteHeader(503)
		return
	}

	if err != nil {
		daemonLogf("%s | ERROR: Could not load configuration file: %s\n", err)
		return
	}

	if conf.Warewulf.Secure {
		if port >= 1024 {
			daemonLogf("DENIED: Connection coming from non-privledged port: %s\n", req.RemoteAddr)
			w.WriteHeader(401)
			return
		}
	}

	node, err := nodes.FindByIpaddr(remote[0])
	if err != nil {
		daemonLogf("WARNING: Could not find node by IP address: %s\n", remote[0])
		w.WriteHeader(404)
		return
	}

	if node.Id.Defined() {
		daemonLogf("REQ:   %15s: %s (unknown/unconfigured node)\n", node.Id.Get(), req.URL.Path)
		w.WriteHeader(404)
		return
	} else {
		daemonLogf("REQ:   %15s: %s\n", node.Id.Get(), req.URL.Path)
	}

	if node.RuntimeOverlay.Defined() {
		fileName := config.RuntimeOverlayImage(node.Id.Get())

		err := sendFile(w, fileName, node.Id.Get())
		if err != nil {
			daemonLogf("ERROR: %s\n", err)
		}
	} else {
		w.WriteHeader(503)
		daemonLogf("WARNING: No 'runtime system-overlay' set for node %s\n", node.Id.Get())
	}
}
