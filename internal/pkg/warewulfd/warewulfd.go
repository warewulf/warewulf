package warewulfd

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/hpcng/warewulf/internal/pkg/warewulfconf"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/pkg/errors"
)

// TODO: https://github.com/danderson/netboot/blob/master/pixiecore/dhcp.go
// TODO: https://github.com/pin/tftp

func RunServer() error {

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP)

	go func() {
		for range c {
			daemonLogf("Recieved SIGHUP, reloading...\n")
			err := LoadNodeDB()
			if err != nil {
				wwlog.Printf(wwlog.ERROR, "Could not load node DB: %s\n", err)
			}

			err = LoadNodeStatus()
			if err != nil {
				wwlog.Printf(wwlog.ERROR, "Could not prepopulate node status DB: %s\n", err)
			}
		}
	}()

	err := LoadNodeDB()
	if err != nil {
		fmt.Printf("ERROR: Could not load database: %s\n", err)
	}

	http.HandleFunc("/ipxe/", IpxeSend)
	http.HandleFunc("/kernel/", KernelSend)
	http.HandleFunc("/kmods/", KmodsSend)
	http.HandleFunc("/container/", ContainerSend)
	http.HandleFunc("/overlay-system/", SystemOverlaySend)
	http.HandleFunc("/overlay-runtime/", RuntimeOverlaySend)
	http.HandleFunc("/status", StatusSend)

	conf, err := warewulfconf.New()
	if err != nil {
		return errors.Wrap(err, "could not get Warewulf configuration")
	}

	daemonPort := conf.Warewulf.Port
	daemonLogf("Starting HTTPD REST service on port %d\n", daemonPort)

	err = http.ListenAndServe(":"+strconv.Itoa(daemonPort), nil)
	if err != nil {
		return errors.Wrap(err, "Could not start listening service")
	}

	return nil
}
