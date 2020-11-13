package warewulfd_responses

import (
	"fmt"
	"github.com/hpcng/warewulf/internal/pkg/assets"
	"github.com/hpcng/warewulf/internal/pkg/config"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"log"
	"net/http"
	"strconv"
	"strings"
	"text/template"
)

type iPxeTemplate struct {
	Hostname		string
	Fqdn           	string
	Vnfs           	string
	Hwaddr		    string
	Ipaddr			string
	Port			string
	Kernelargs		string
}


func IpxeSend(w http.ResponseWriter, req *http.Request) {
	url := strings.Split(req.URL.Path, "/")

	if url[2] == "" {
		log.Printf("ERROR: Bad iPXE request from %s\n", req.RemoteAddr)
		return
	}

	hwaddr := strings.ReplaceAll(url[2], "-", ":")
	node, err := assets.FindByHwaddr(hwaddr)
	if err != nil {
		log.Printf("Could not find HW Addr: %s: %s\n", hwaddr, err)
		w.WriteHeader(404)
		return
	}

	if node.HostName != "" {
		conf := config.New()

		log.Printf("IPXE:  %15s: %s\n", node.Fqdn, req.URL.Path)

		// TODO: Fix template path to use config package
		ipxeTemplate := fmt.Sprintf("/etc/warewulf/ipxe/%s.ipxe", node.Ipxe)

		tmpl, err := template.ParseFiles(ipxeTemplate)
		if err != nil {
			wwlog.Printf(wwlog.ERROR, "%s\n", err)
			return
		}

		var replace iPxeTemplate

		replace.Fqdn = node.Fqdn
		replace.Ipaddr = conf.Ipaddr
		replace.Port = strconv.Itoa(conf.Port)
		replace.Hostname = node.HostName
		replace.Hwaddr = url[2]
		replace.Vnfs = node.Vnfs
		replace.Kernelargs = node.KernelArgs

		err = tmpl.Execute(w, replace)
		if err != nil {
			wwlog.Printf(wwlog.ERROR, "%s\n", err)
			return
		}

		log.Printf("SEND:  %15s: %s\n", node.Fqdn, ipxeTemplate)

	} else {
		log.Printf("ERROR: iPXE request from unknown Node (hwaddr=%s)\n", url[2])
	}
	return
}
