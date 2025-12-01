package server

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path"
	"strconv"
	"strings"
	"syscall"

	warewulfconf "github.com/warewulf/warewulf/internal/pkg/config"
	"github.com/warewulf/warewulf/internal/pkg/util"
	"github.com/warewulf/warewulf/internal/pkg/warewulfd"
	"github.com/warewulf/warewulf/internal/pkg/warewulfd/api"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

// TODO: https://github.com/danderson/netboot/blob/master/pixiecore/dhcp.go
// TODO: https://github.com/pin/tftp
/*
wrapper type for the server mux as shim requests http://efiboot//grub.efi
which is filtered out by http to `301 Moved Permanently` what
shim.efi can't handle. So filter out `//` before they hit go/http.
Makes go/http more to behave like apache
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

func defaultHandler() *slashFix {
	var wwHandler http.ServeMux
	wwHandler.HandleFunc("/provision/", warewulfd.ProvisionSend)
	wwHandler.HandleFunc("/ipxe/", warewulfd.ProvisionSend)
	wwHandler.HandleFunc("/efiboot/", warewulfd.ProvisionSend)
	wwHandler.HandleFunc("/kernel/", warewulfd.ProvisionSend)
	wwHandler.HandleFunc("/container/", warewulfd.ProvisionSend)
	wwHandler.HandleFunc("/overlay-system/", warewulfd.ProvisionSend)
	wwHandler.HandleFunc("/overlay-runtime/", warewulfd.ProvisionSend)
	wwHandler.HandleFunc("/overlay-file/", warewulfd.OverlaySend)
	wwHandler.HandleFunc("/status", warewulfd.StatusSend)
	return &slashFix{&wwHandler}
}

func RunServer() error {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP)

	go func() {
		for range c {
			wwlog.Info("Received SIGHUP, reloading...")
			warewulfd.Reload()
		}
	}()

	warewulfd.Reload()

	conf := warewulfconf.Get()
	daemonPort := conf.Warewulf.Port

	auth := warewulfconf.NewAuthentication()
	if util.IsFile(conf.Paths.AuthenticationConf()) {
		if err := auth.Read(conf.Paths.AuthenticationConf()); err != nil {
			wwlog.Warn("%w\n", err)
		}
	}

	apiHandler := api.Handler(auth, conf.API.AllowedIPNets())
	defaultHandler := defaultHandler()
	dispatchHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api") && conf.API != nil && conf.API.Enabled() {
			apiHandler.ServeHTTP(w, r)
		} else {
			defaultHandler.ServeHTTP(w, r)
		}
	})

	if conf.Warewulf.EnableHttps() {
		key := path.Join(conf.Paths.Sysconfdir, "warewulf", "keys", "warewulf.key")
		crt := path.Join(conf.Paths.Sysconfdir, "warewulf", "keys", "warewulf.crt")

		if !util.IsFile(key) || !util.IsFile(crt) {
			wwlog.Error("HTTPS enabled but keys not found in %s", path.Join(conf.Paths.Sysconfdir, "warewulf", "keys"))
		} else {
			go func() {
				wwlog.Info("Starting HTTPS service on port %d", conf.Warewulf.SecurePort)
				if err := http.ListenAndServeTLS(":"+strconv.Itoa(conf.Warewulf.SecurePort), crt, key, dispatchHandler); err != nil {
					wwlog.Error("Could not start HTTPS service: %s", err)
				}
			}()
		}
	}

	wwlog.Info("Starting HTTP service on port %d", daemonPort)
	if err := http.ListenAndServe(":"+strconv.Itoa(daemonPort), dispatchHandler); err != nil {
		return fmt.Errorf("could not start listening service: %w", err)
	}

	return nil
}
