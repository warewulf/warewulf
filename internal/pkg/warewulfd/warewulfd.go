package warewulfd

import (
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	warewulfconf "github.com/hpcng/warewulf/internal/pkg/config"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// TODO: https://github.com/danderson/netboot/blob/master/pixiecore/dhcp.go
// TODO: https://github.com/pin/tftp
/*
wrapper type for the server mux as him requests http://efiboot//grub.efi which is filtered out by http to `301 Moved Permanently` which
which shim.fi can't handle. So filter out `//` before they hit http
*/
type slashFix struct {
	mux http.Handler
}

/*
Filter out the '//'
*/
func (h *slashFix) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	r.URL.Path = strings.Replace(r.URL.Path, "//", "/", -1)
	h.mux.ServeHTTP(w, r)
}

func RunServer() error {
	err := DaemonInitLogging()
	if err != nil {
		return errors.Wrap(err, "Failed to initialize logging")
	}

	conf := warewulfconf.Get()
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
			if conf.Warewulf.EnablePrometheus {
				UpdateContainerSize()
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

	err = CopyShimGrub()
	if err != nil {
		wwlog.Warn("couldn't copy default shim: %s", err)
	}
	var wwHandler http.ServeMux
	wwHandler.HandleFunc("/provision/", ProvisionSend)
	wwHandler.HandleFunc("/ipxe/", ProvisionSend)
	wwHandler.HandleFunc("/efiboot/", ProvisionSend)
	wwHandler.HandleFunc("/kernel/", ProvisionSend)
	wwHandler.HandleFunc("/kmods/", ProvisionSend)
	wwHandler.HandleFunc("/container/", ProvisionSend)
	wwHandler.HandleFunc("/overlay-system/", ProvisionSend)
	wwHandler.HandleFunc("/overlay-runtime/", ProvisionSend)
	wwHandler.HandleFunc("/status", StatusSend)
	wwHandler.HandleFunc("/sentstatus", SentStatus)
	if conf.Warewulf.EnablePrometheus {
		go func() {
			for {
				// update the container sizes every two h
				UpdateContainerSize()
				time.Sleep(3600 * 2 * time.Millisecond)
			}
		}()
		prometheus.MustRegister(NewCollector())
		wwHandler.Handle("/metrics", promhttp.Handler())
	}

	daemonPort := conf.Warewulf.Port
	wwlog.Serv("Starting HTTPD REST service on port %d", daemonPort)

	err = http.ListenAndServe(":"+strconv.Itoa(daemonPort), &slashFix{&wwHandler})
	if err != nil {
		return errors.Wrap(err, "Could not start listening service")
	}

	return nil
}
