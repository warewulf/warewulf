package warewulfd

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"text/template"

	"github.com/hpcng/warewulf/internal/pkg/node"
	"github.com/hpcng/warewulf/internal/pkg/overlay"
	"github.com/hpcng/warewulf/internal/pkg/warewulfconf"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
)

type iPxeTemplate struct {
	Message       string
	WaitTime      string
	Hostname      string
	Fqdn          string
	Id            string
	Cluster       string
	ContainerName string
	Hwaddr        string
	Ipaddr        string
	Port          string
	KernelArgs    string
	KernelVersion string
}

func IpxeSend(w http.ResponseWriter, req *http.Request) {

	url := strings.Split(req.URL.Path, "/")
	var unconfiguredNode bool

	if url[2] == "" {
		daemonLogf("ERROR: Bad iPXE request from %s\n", req.RemoteAddr)
		w.WriteHeader(404)
		return
	}

	hwaddr := strings.ReplaceAll(url[2], "-", ":")

	conf, err := warewulfconf.New()
	if err != nil {
		daemonLogf("ERROR: Could not open Warewulf configuration: %s\n", err)
		w.WriteHeader(503)
		return
	}

	nodeobj, err := GetNode(hwaddr)

	if err != nil {
		// If we failed to find a node, let's see if we can add one...
		var netdev string

		nodeDB, err := node.New()
		if err != nil {
			daemonLogf("Could not read node configuration file: %s\n", err)
			w.WriteHeader(503)
			return
		}

		daemonLogf("IPXEREQ:   %s (node not configured)\n", hwaddr)

		n, netdev, err := nodeDB.FindDiscoverableNode()
		if err != nil {
			unconfiguredNode = true

		} else {
			n.NetDevs[netdev].Hwaddr.Set(hwaddr)
			n.Discoverable.SetB(false)
			err := nodeDB.NodeUpdate(n)
			if err != nil {
				daemonLogf("IPXEREQ:   %s (failed to set node configuration)\n", hwaddr)

				unconfiguredNode = true
			} else {
				err := nodeDB.Persist()
				if err != nil {
					daemonLogf("IPXEREQ:   %s (failed to persist node configuration)\n", hwaddr)

					unconfiguredNode = true
				} else {
					nodeobj = n
					_ = overlay.BuildSystemOverlay([]node.NodeInfo{n})
					_ = overlay.BuildRuntimeOverlay([]node.NodeInfo{n})

					daemonLogf("IPXEREQ:   %s (node automatically configured)\n", hwaddr)

					err := LoadNodeDB()
					if err != nil {
						daemonLogf("Could not reload configuration: %s\n", err)
					}

				}
			}
		}
	}

	if unconfiguredNode {
		daemonLogf("IPXEREQ:   %s (unknown/unconfigured node)\n", hwaddr)

		tmpl, err := template.ParseFiles("/etc/warewulf/ipxe/unconfigured.ipxe")
		if err != nil {
			daemonLogf("ERROR: Could not parse unconfigured node IPXE template: %s\n", err)
			return
		}

		var replace iPxeTemplate

		replace.Hwaddr = hwaddr

		err = tmpl.Execute(w, replace)
		if err != nil {
			daemonLogf("ERROR: Could not update unconfigured node IPXE template: %s\n", err)
			return
		}

		return

	} else {

		ipxeTemplate := fmt.Sprintf("/etc/warewulf/ipxe/%s.ipxe", nodeobj.Ipxe.Get())

		tmpl, err := template.ParseFiles(ipxeTemplate)
		if err != nil {
			wwlog.Printf(wwlog.ERROR, "%s\n", err)
			return
		}

		var replace iPxeTemplate

		replace.Id = nodeobj.Id.Get()
		replace.Cluster = nodeobj.ClusterName.Get()
		replace.Fqdn = nodeobj.Id.Get()
		replace.Ipaddr = conf.Ipaddr
		replace.Port = strconv.Itoa(conf.Warewulf.Port)
		replace.Hostname = nodeobj.Id.Get()
		replace.Hwaddr = url[2]
		replace.ContainerName = nodeobj.ContainerName.Get()
		replace.KernelArgs = nodeobj.KernelArgs.Get()
		replace.KernelVersion = nodeobj.KernelVersion.Get()

		err = tmpl.Execute(w, replace)
		if err != nil {
			wwlog.Printf(wwlog.ERROR, "%s\n", err)
			return
		}

		daemonLogf("SEND:  %15s: %s\n", nodeobj.Id.Get(), ipxeTemplate)

	}
}
