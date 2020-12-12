package warewulfd

import (
	"fmt"
	"github.com/hpcng/warewulf/internal/pkg/node"
	"github.com/hpcng/warewulf/internal/pkg/overlay"
	"github.com/hpcng/warewulf/internal/pkg/warewulfconf"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"log"
	"net/http"
	"strconv"
	"strings"
	"text/template"
)

type iPxeTemplate struct {
	Message       string
	WaitTime      string
	Hostname      string
	Fqdn          string
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

	nodeDB, err := node.New()
	if err != nil {
		log.Printf("Could not read node configuration file: %s\n", err)
		w.WriteHeader(503)
		return
	}

	if url[2] == "" {
		log.Printf("ERROR: Bad iPXE request from %s\n", req.RemoteAddr)
		w.WriteHeader(404)
		return
	}

	conf, err := warewulfconf.New()
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "%s\n", err)
		w.WriteHeader(503)
		return
	}

	hwaddr := strings.ReplaceAll(url[2], "-", ":")

	n, err := nodeDB.FindByHwaddr(hwaddr)
	if err != nil {
		// If we failed to find a node, let's see if we can add one...
		var netdev string

		wwlog.Printf(wwlog.INFO, "Node was not found, looking for discoverable nodes...\n")

		n, netdev, err = nodeDB.FindUnconfiguredNode()
		if err != nil {
			wwlog.Printf(wwlog.WARN, "Node was not found, no nodes are discoverable...\n")
			unconfiguredNode = true

		} else {
			wwlog.Printf(wwlog.INFO, "Adding new configuration to discoverable node: %s\n", n.Id.Get())

			n.NetDevs[netdev].Hwaddr.Set(hwaddr)
			err := nodeDB.NodeUpdate(n)
			if err != nil {
				wwlog.Printf(wwlog.ERROR, "Could not add discovered configuration for node: %s\n", n.Id.Get())
				unconfiguredNode = true
			} else {
				err := nodeDB.Persist()
				if err != nil {
					wwlog.Printf(wwlog.ERROR, "Could not persist new node configuration while adding node: %s\n", n.Id.Get())
					unconfiguredNode = true
				} else {
					_ = overlay.BuildSystemOverlay([]node.NodeInfo{n})
					_ = overlay.BuildRuntimeOverlay([]node.NodeInfo{n})
				}
			}
		}
	}

	if unconfiguredNode == true {
		log.Printf("UNCONFIGURED NODE:  %15s\n", hwaddr)

		ipxeTemplate := fmt.Sprintf("/etc/warewulf/ipxe/unconfigured.ipxe", n.Ipxe.Get())

		tmpl, err := template.ParseFiles(ipxeTemplate)
		if err != nil {
			wwlog.Printf(wwlog.ERROR, "%s\n", err)
			return
		}

		var replace iPxeTemplate

		replace.Hwaddr = hwaddr

		err = tmpl.Execute(w, replace)
		if err != nil {
			wwlog.Printf(wwlog.ERROR, "%s\n", err)
			return
		}

		return

	} else {
		log.Printf("IPXE:  %15s: %s\n", n.Id.Get(), req.URL.Path)

		ipxeTemplate := fmt.Sprintf("/etc/warewulf/ipxe/%s.ipxe", n.Ipxe.Get())

		tmpl, err := template.ParseFiles(ipxeTemplate)
		if err != nil {
			wwlog.Printf(wwlog.ERROR, "%s\n", err)
			return
		}

		var replace iPxeTemplate

		replace.Fqdn = n.Id.Get()
		replace.Ipaddr = conf.Ipaddr
		replace.Port = strconv.Itoa(conf.Warewulf.Port)
		replace.Hostname = n.Id.Get()
		replace.Hwaddr = url[2]
		replace.ContainerName = n.ContainerName.Get()
		replace.KernelArgs = n.KernelArgs.Get()
		replace.KernelVersion = n.KernelVersion.Get()

		err = tmpl.Execute(w, replace)
		if err != nil {
			wwlog.Printf(wwlog.ERROR, "%s\n", err)
			return
		}

		log.Printf("SEND:  %15s: %s\n", n.Id.Get(), ipxeTemplate)

	}
	return
}
