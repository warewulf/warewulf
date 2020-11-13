package warewulfd_responses

import (
	"github.com/hpcng/warewulf/internal/pkg/config"
	"github.com/hpcng/warewulf/internal/pkg/vnfs"
	"log"
	"net/http"
)

func VnfsSend(w http.ResponseWriter, req *http.Request) {
	config := config.New()

	node, err := getSanity(req)
	if err != nil {
		w.WriteHeader(404)
		log.Panicln(err)
		return
	}

	if node.Vnfs != "" {
		v := vnfs.New(node.Vnfs)

		err := sendFile(w, config.VnfsImage(v.NameClean()), node.Fqdn)
		if err != nil {
			log.Printf("ERROR: %s\n", err)
		} else {
			log.Printf("SEND:  %15s: %s\n", node.Fqdn, config.VnfsImage(v.NameClean()))
		}
	} else {
		w.WriteHeader(503)
		log.Printf("ERROR: No VNFS set for node %s\n", node.Fqdn)
	}

	return
}
