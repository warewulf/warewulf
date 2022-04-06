package warewulfd

import (
	"net/http"
	"path"
	"strconv"
	"strings"
	"text/template"

	"github.com/hpcng/warewulf/internal/pkg/buildconfig"
	nodepkg "github.com/hpcng/warewulf/internal/pkg/node"
	"github.com/hpcng/warewulf/internal/pkg/overlay"
	"github.com/hpcng/warewulf/internal/pkg/warewulfconf"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
)

type iPxeTemplate struct {
	Message        string
	WaitTime       string
	Hostname       string
	Fqdn           string
	Id             string
	Cluster        string
	ContainerName  string
	Hwaddr         string
	Ipaddr         string
	Port           string
	KernelArgs     string
	KernelOverride string
}

func IpxeSend(w http.ResponseWriter, req *http.Request) {
	conf, err := warewulfconf.New()
	if err != nil {
		daemonLogf("ERROR: Could not open Warewulf configuration: %s\n", err)
		w.WriteHeader(503)
		return
	}

	rinfo, err := parseReq(req)
	if err != nil {
		w.WriteHeader(404)
		daemonLogf("ERROR: %s\n", err)
		return
	}

	node, err := GetNode(rinfo.hwaddr)
	if err != nil {
		// If we failed to find a node, let's see if we can add one...
		var netdev string
		var unconfiguredNode bool

		daemonLogf("IPXEREQ:   %s (node not configured)\n", rinfo.hwaddr)

		nodeDB, err := nodepkg.New()
		if err != nil {
			daemonLogf("Could not read node configuration file: %s\n", err)
			w.WriteHeader(503)
			return
		}

		n, netdev, err := nodeDB.FindDiscoverableNode()
		if err != nil {
			unconfiguredNode = true

		} else {
			n.NetDevs[netdev].Hwaddr.Set(rinfo.hwaddr)
			n.Discoverable.SetB(false)
			err := nodeDB.NodeUpdate(n)
			if err != nil {
				daemonLogf("IPXEREQ:   %s (failed to set node configuration)\n", rinfo.hwaddr)

				unconfiguredNode = true
			} else {
				err := nodeDB.Persist()
				if err != nil {
					daemonLogf("IPXEREQ:   %s (failed to persist node configuration)\n", rinfo.hwaddr)

					unconfiguredNode = true
				} else {
					node = n
					_ = overlay.BuildAllOverlays([]nodepkg.NodeInfo{n})

					daemonLogf("IPXEREQ:   %s (node automatically configured)\n", rinfo.hwaddr)

					err := LoadNodeDB()
					if err != nil {
						daemonLogf("Could not reload configuration: %s\n", err)
					}

				}
			}
		}
		if unconfiguredNode {
			daemonLogf("IPXEREQ:   %s (unknown/unconfigured node)\n", rinfo.hwaddr)

			tmpl, err := template.ParseFiles(path.Join(buildconfig.SYSCONFDIR(), "/warewulf/ipxe/unconfigured.ipxe"))
			if err != nil {
				daemonLogf("ERROR: Could not parse unconfigured node IPXE template: %s\n", err)
				return
			}

			var replace iPxeTemplate

			replace.Hwaddr = rinfo.hwaddr

			err = tmpl.Execute(w, replace)
			if err != nil {
				daemonLogf("ERROR: Could not update unconfigured node IPXE template: %s\n", err)
				return
			}

			return
		}
	}

	if node.AssetKey.Defined() && node.AssetKey.Get() != rinfo.assetkey {
		w.WriteHeader(404)
		daemonLogf("ERROR: Incorrect asset key for node: %s\n", node.Id.Get())
		updateStatus(node.Id.Get(), "IPXE", "BAD_ASSET", rinfo.ipaddr)
		return
	}

	ipxeTemplate := path.Join(buildconfig.SYSCONFDIR(), "warewulf/ipxe/"+node.Ipxe.Get()+".ipxe")

	tmpl, err := template.ParseFiles(ipxeTemplate)
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "%s\n", err)
		return
	}

	var replace iPxeTemplate

	replace.Id = node.Id.Get()
	replace.Cluster = node.ClusterName.Get()
	replace.Fqdn = node.Id.Get()
	replace.Ipaddr = conf.Ipaddr
	replace.Port = strconv.Itoa(conf.Warewulf.Port)
	replace.Hostname = node.Id.Get()
	replace.Hwaddr = rinfo.hwaddr
	replace.ContainerName = node.ContainerName.Get()
	replace.KernelArgs = node.Kernel.Args.Get()
	replace.KernelOverride = node.Kernel.Override.Get()

	err = tmpl.Execute(w, replace)
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "%s\n", err)
		return
	}

	daemonLogf("SEND:  %15s: %s\n", node.Id.Get(), ipxeTemplate)

	updateStatus(node.Id.Get(), "IPXE", node.Ipxe.Get()+".ipxe", strings.Split(req.RemoteAddr, ":")[0])

}
