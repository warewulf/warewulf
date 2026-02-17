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

func configureHandler(includeRuntime bool, apiHandler http.Handler) *slashFix {
	var wwHandler http.ServeMux
	wwHandler.HandleFunc("/provision/", warewulfd.ProvisionSend)
	wwHandler.HandleFunc("/ipxe/", warewulfd.ProvisionSend)
	wwHandler.HandleFunc("/efiboot/", warewulfd.ProvisionSend)
	wwHandler.HandleFunc("/kernel/", warewulfd.ProvisionSend)
	wwHandler.HandleFunc("/container/", warewulfd.ProvisionSend)
	wwHandler.HandleFunc("/overlay-system/", warewulfd.ProvisionSend)
	if includeRuntime {
		wwHandler.HandleFunc("/overlay-runtime/", warewulfd.ProvisionSend)
	}
	wwHandler.HandleFunc("/overlay-file/", warewulfd.OverlaySend)
	wwHandler.HandleFunc("/status", warewulfd.StatusSend)

	if apiHandler != nil {
		wwHandler.Handle("/api/", apiHandler)
	}

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

	var apiHandler http.Handler
	if conf.API != nil && conf.API.Enabled() {
		apiHandler = api.Handler(auth, conf.API.AllowedIPNets())
	}

	httpHandler := configureHandler(!conf.Warewulf.EnableTLS(), apiHandler)

	errChan := make(chan error, 2)

	if conf.Warewulf.EnableTLS() {
		key := path.Join(conf.Paths.Sysconfdir, "warewulf", "tls", "warewulf.key")
		crt := path.Join(conf.Paths.Sysconfdir, "warewulf", "tls", "warewulf.crt")

		if !util.IsFile(key) || !util.IsFile(crt) {
			return fmt.Errorf("TLS enabled but keys not found in %s, run 'wwctl configure tls --create' to generate keys", path.Join(conf.Paths.Sysconfdir, "warewulf", "tls"))
		}
		httpsHandler := configureHandler(true, apiHandler)
		go func() {
			wwlog.Info("Starting HTTPS service on port %d", conf.Warewulf.SecurePort)
			if err := http.ListenAndServeTLS(":"+strconv.Itoa(conf.Warewulf.SecurePort), crt, key, httpsHandler); err != nil {
				errChan <- fmt.Errorf("could not start HTTPS service: %w", err)
			}
		}()
	}

	go func() {
		wwlog.Info("Starting HTTP service on port %d", daemonPort)
		if err := http.ListenAndServe(":"+strconv.Itoa(daemonPort), httpHandler); err != nil {
			errChan <- fmt.Errorf("could not start HTTP service: %w", err)
		}
	}()

	return <-errChan
}
