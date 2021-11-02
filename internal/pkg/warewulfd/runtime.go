package warewulfd

import (
	"github.com/hpcng/warewulf/internal/pkg/config"
	"github.com/hpcng/warewulf/internal/pkg/node"
	"github.com/hpcng/warewulf/internal/pkg/overlay"
	"github.com/hpcng/warewulf/internal/pkg/util"
	"github.com/hpcng/warewulf/internal/pkg/warewulfconf"
	"net/http"
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
		daemonLogf("ERROR: Could not read node configuration file: %s\n", err)
		w.WriteHeader(503)
		return
	}

	// save way to obtain IP and Port, for IPv4 and IPv6
	host, port, err := getHostPort(w, req)
	if err != nil {
		daemonLogf("ERROR: failed to obtain host and port: %s\n", err)
		return
	}

	if conf.Warewulf.Secure {
		if port >= 1024 {
			daemonLogf("DENIED: Connection coming from non-privileged port: %s\n", req.RemoteAddr)
			w.WriteHeader(401)
			return
		}
	}

	var n node.NodeInfo
	// default to IP based host identification
	if !conf.Warewulf.MacIdentify {
		n, err = nodes.FindByIpaddr(host)
		if err != nil {
			daemonLogf("WARNING: Could not find node by IP address: %s\n", host)
			w.WriteHeader(404)
			return
		}
	} else {
		hwAddresses, ok := req.URL.Query()["hwAddr"]
		if !ok || len(hwAddresses[0]) < 1 {
			daemonLogf("ERROR: Url Param 'hwAddr' is missing")
			return
		}
		for _, hwa := range hwAddresses {
			n, err = nodes.FindByHwaddr(hwa)
			if n.Id.Defined() {
				daemonLogf("DEBUG: nodeId: %s, HardwareAddr: %s\n", n.Id.Get(), hwa)
				break
			}
		}
	}

	if !n.Id.Defined() {
		daemonLogf("REQ:   %15s: %s (unknown/unconfigured node)\n", n.Id.Get(), req.URL.Path)
		w.WriteHeader(404)
		return
	} else {
		daemonLogf("REQ:   %15s: %s\n", n.Id.Get(), req.URL.Path)
	}

	if n.RuntimeOverlay.Defined() {
		fileName := config.RuntimeOverlayImage(n.Id.Get())

		if conf.Warewulf.AutobuildOverlays {
			if !util.IsFile(fileName) || util.PathIsNewer(fileName, node.ConfigFile) || util.PathIsNewer(fileName, config.RuntimeOverlaySource(n.RuntimeOverlay.Get())) {
				daemonLogf("BUILD: %15s: Runtime Overlay\n", n.Id.Get())
				_ = overlay.BuildRuntimeOverlay([]node.NodeInfo{n})
			}
		}

		http.ServeFile(w, req, fileName)
		if err != nil {
			daemonLogf("ERROR: %s\n", err)
		}
	} else {
		w.WriteHeader(503)
		daemonLogf("WARNING: No 'runtime system-overlay' set for node %s\n", n.Id.Get())
	}
}
