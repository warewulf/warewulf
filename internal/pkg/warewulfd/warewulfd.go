package warewulfd

import (
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
	err := DaemonInitLogging()
	if err != nil {
		return errors.Wrap(err, "Failed to initialize logging")
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP)

	go func() {
		for range c {
			wwlog.Warn("Received SIGHUP, reloading...")
			err := LoadNodeDB()
			if err != nil {
				wwlog.Error("Could not load node DB: %s", err)
			}

			err = LoadNodeStatus()
			if err != nil {
				wwlog.Error("Could not prepopulate node status DB: %s", err)
			}
		}
	}()

	err = LoadNodeDB()
	if err != nil {
		wwlog.Error("Could not load database: %s", err)
	}

	err = LoadNodeStatus()
	if err != nil {
		wwlog.Error("Could not prepopulate node status DB: %s", err)
	}

	http.HandleFunc("/provision/", ProvisionSend)
	http.HandleFunc("/ipxe/", ProvisionSend)
	http.HandleFunc("/kernel/", ProvisionSend)
	http.HandleFunc("/kmods/", ProvisionSend)
	http.HandleFunc("/container/", ProvisionSend)
	http.HandleFunc("/overlay-system/", ProvisionSend)
	http.HandleFunc("/overlay-runtime/", ProvisionSend)
	http.HandleFunc("/status", StatusSend)

	conf := warewulfconf.New()

	daemonPort := conf.Warewulf.Port
	wwlog.Serv("Starting HTTPD REST service on port %d", daemonPort)

	err = http.ListenAndServe(":"+strconv.Itoa(daemonPort), nil)
	if err != nil {
		return errors.Wrap(err, "Could not start listening service")
	}

	return nil
}
