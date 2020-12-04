package response

import (
	"fmt"
	"github.com/hpcng/warewulf/internal/pkg/node"
	"github.com/hpcng/warewulf/internal/pkg/warewulfconf"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"log"
	"net/http"
	"strconv"
	"strings"
	"text/template"
)

type iPxeTemplate struct {
	Hostname   string
	Fqdn       string
	Vnfs       string
	Hwaddr     string
	Ipaddr     string
	Port       string
	Kernelargs string
}

func IpxeSend(w http.ResponseWriter, req *http.Request) {
	url := strings.Split(req.URL.Path, "/")

	nodes, err := node.New()
	if err != nil {
		log.Printf("Could not read node configuration file: %s\n", err)
		w.WriteHeader(503)
		return
	}

	if url[2] == "" {
		log.Printf("ERROR: Bad iPXE request from %s\n", req.RemoteAddr)
		return
	}

	hwaddr := strings.ReplaceAll(url[2], "-", ":")
	node, err := nodes.FindByHwaddr(hwaddr)
	if err != nil {
		log.Printf("Could not find HW Addr: %s: %s\n", hwaddr, err)
		w.WriteHeader(404)
		return
	}

	if node.Id.Defined() == true {
		conf, err := warewulfconf.New()
		if err != nil {
			wwlog.Printf(wwlog.ERROR, "%s\n", err)
			return
		}

		log.Printf("IPXE:  %15s: %s\n", node.Id.Get(), req.URL.Path)

		// TODO: Fix template path to use config package
		ipxeTemplate := fmt.Sprintf("/etc/warewulf/ipxe/%s.ipxe", node.Ipxe.Get())

		tmpl, err := template.ParseFiles(ipxeTemplate)
		if err != nil {
			wwlog.Printf(wwlog.ERROR, "%s\n", err)
			return
		}

		var replace iPxeTemplate

		replace.Fqdn = node.Id.Get()
		replace.Ipaddr = conf.Ipaddr
		replace.Port = strconv.Itoa(conf.Warewulf.Port)
		replace.Hostname = node.Id.Get()
		replace.Hwaddr = url[2]
		replace.Vnfs = node.Vnfs.Get()
		replace.Kernelargs = node.KernelArgs.Get()

		err = tmpl.Execute(w, replace)
		if err != nil {
			wwlog.Printf(wwlog.ERROR, "%s\n", err)
			return
		}

		log.Printf("SEND:  %15s: %s\n", node.Id.Get(), ipxeTemplate)

	} else {
		log.Printf("ERROR: iPXE request from unknown Node (hwaddr=%s)\n", url[2])
	}
	return
}
