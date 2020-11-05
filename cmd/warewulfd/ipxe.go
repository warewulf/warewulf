package main

import (
	"bufio"
	"fmt"
	"github.com/hpcng/warewulf/internal/pkg/assets"
	"github.com/hpcng/warewulf/internal/pkg/config"
	"log"
	"net/http"
	"os"
	"strconv"
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
		replace := make(map[string]string)

		conf, err := config.New()
		if err != nil {
			log.Printf("Could not get config: %s\n", err)
			return
		}

		ipxeTemplate := fmt.Sprintf("/etc/warewulf/ipxe/%s.ipxe", node.Ipxe)
		sourceFD, err := os.Open(ipxeTemplate)
		if err != nil {
			log.Printf("ERROR: Could not open iPXE Template: %s\n", err)
			w.WriteHeader(404)
			return
		}

		scanner := bufio.NewScanner(sourceFD)

		for scanner.Scan() {
			newLine := scanner.Text()

			newLine = strings.ReplaceAll(newLine, "@HWADDR@", url[2])
			newLine = strings.ReplaceAll(newLine, "@IPADDR@", conf.Ipaddr)
			newLine = strings.ReplaceAll(newLine, "@HOSTNAME@", node.HostName)
			newLine = strings.ReplaceAll(newLine, "@PORT@", strconv.Itoa(conf.Port))

			fmt.Fprintln(w, newLine)
		}

	} else {
		log.Printf("ERROR: iPXE request from unknown Node (hwaddr=%s)\n", url[2])
	}
	return
}
