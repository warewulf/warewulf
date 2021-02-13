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

	if url[2] == "" {
		log.Printf("ERROR: Bad iPXE request from %s\n", req.RemoteAddr)
		w.WriteHeader(404)
		return
	}

	hwaddr := strings.ReplaceAll(url[2], "-", ":")

	conf, err := warewulfconf.New()
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "%s\n", err)
		w.WriteHeader(503)
		return
	}

	nodeobj, err := GetNode(hwaddr)

	if err != nil {
		// If we failed to find a node, let's see if we can add one...
		var netdev string

		nodeDB, err := node.New()
		if err != nil {
			log.Printf("Could not read node configuration file: %s\n", err)
			w.WriteHeader(503)
			return
		}

		wwlog.Printf(wwlog.VERBOSE, "Node was not found, looking for discoverable nodes...\n")

		n, netdev, err := nodeDB.FindDiscoverableNode()
		if err != nil {
			wwlog.Printf(wwlog.WARN, "No nodes are set as discoverable...\n")
			unconfiguredNode = true

		} else {
			wwlog.Printf(wwlog.INFO, "Adding new configuration to discoverable node: %s\n", n.Id.Get())

			n.NetDevs[netdev].Hwaddr.Set(hwaddr)
			n.Discoverable.SetB(false)
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
					nodeobj = n
					wwlog.Printf(wwlog.INFO, "Building System Overlay:\n")
					_ = overlay.BuildSystemOverlay([]node.NodeInfo{n})
					wwlog.Printf(wwlog.INFO, "Building Runtime Overlay:\n")
					_ = overlay.BuildRuntimeOverlay([]node.NodeInfo{n})
				}
			}
		}
	}

	if unconfiguredNode == true {
		log.Printf("UNCONFIGURED NODE:  %15s\n", hwaddr)

		tmpl, err := template.ParseFiles("/etc/warewulf/ipxe/unconfigured.ipxe")
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

		log.Printf("IPXE:  %15s: %s\n", nodeobj.Id.Get(), req.URL.Path)

		ipxeTemplate := fmt.Sprintf("/etc/warewulf/ipxe/%s.ipxe", nodeobj.Ipxe.Get())

		tmpl, err := template.ParseFiles(ipxeTemplate)
		if err != nil {
			wwlog.Printf(wwlog.ERROR, "%s\n", err)
			return
		}

		var replace iPxeTemplate

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

		log.Printf("SEND:  %15s: %s\n", nodeobj.Id.Get(), ipxeTemplate)

	}
	return
}
