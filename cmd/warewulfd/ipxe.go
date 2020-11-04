package main

import (
	"fmt"
	"github.com/hpcng/warewulf/internal/pkg/assets"
	"log"
	"net/http"
	"strings"
)

func ipxe(w http.ResponseWriter, req *http.Request) {
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
		log.Printf("IPXE:  %15s: hwaddr=%s\n", node.Fqdn, hwaddr)

		fmt.Fprintf(w, "#!ipxe\n")

		fmt.Fprintf(w, "echo Now booting Warewulf - v4 Proof of Concept\n")
		fmt.Fprintf(w, "set base http://192.168.1.1:9873/\n")
		fmt.Fprintf(w, "kernel ${base}/kernel/%s crashkernel=no quiet\n", url[2])
		fmt.Fprintf(w, "initrd ${base}/vnfs/%s\n", url[2])
		fmt.Fprintf(w, "initrd ${base}/kmods/%s\n", url[2])
		fmt.Fprintf(w, "initrd ${base}/overlay-system/%s\n", url[2])
		fmt.Fprintf(w, "boot\n")
	} else {
		log.Printf("ERROR: iPXE request from unknown Node (hwaddr=%s)\n", url[2])
	}
	return
}
