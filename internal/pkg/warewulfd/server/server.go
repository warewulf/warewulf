// Package server starts the Warewulf HTTP(S) server, registers provisioning
// and API routes, and handles TLS configuration and SIGHUP-triggered reloads.
//
// See userdocs/server/routes.rst for more information.
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

func requireTLS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.TLS == nil {
			wwlog.Denied("API request over insecure connection")
			w.WriteHeader(http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func configureRootHandler(apiHandler http.Handler) *slashFix {
	var wwHandler http.ServeMux
	wwHandler.HandleFunc("/provision/", warewulfd.HandleProvision)
	wwHandler.HandleFunc("/ipxe/", warewulfd.HandleIpxe)
	wwHandler.HandleFunc("/efiboot/", warewulfd.HandleEfiBoot)
	wwHandler.HandleFunc("/grub/", warewulfd.HandleGrub)
	wwHandler.HandleFunc("/kernel/", warewulfd.HandleKernel)
	wwHandler.HandleFunc("/image/", warewulfd.HandleImage)
	wwHandler.HandleFunc("/initramfs/", warewulfd.HandleInitramfs)
	wwHandler.HandleFunc("/system/", warewulfd.HandleSystemOverlay)
	wwHandler.HandleFunc("/runtime/", warewulfd.HandleRuntimeOverlay)
	wwHandler.HandleFunc("/status", warewulfd.HandleStatus)

	/* Deprecated */
	wwHandler.HandleFunc("/container/", warewulfd.HandleImage)
	wwHandler.HandleFunc("/overlay-system/", warewulfd.HandleSystemOverlay)
	wwHandler.HandleFunc("/overlay-runtime/", warewulfd.HandleRuntimeOverlay)
	wwHandler.HandleFunc("/overlay-file/", warewulfd.HandleOverlayFile)

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
		if conf.API.TLSEnabled() {
			apiHandler = requireTLS(apiHandler)
		}
	}

	httpHandler := configureRootHandler(apiHandler)

	errChan := make(chan error, 2)

	if conf.Warewulf.TLSEnabled() {
		key := path.Join(conf.Paths.Sysconfdir, "warewulf", "tls", "warewulf.key")
		crt := path.Join(conf.Paths.Sysconfdir, "warewulf", "tls", "warewulf.crt")

		if !util.IsFile(key) || !util.IsFile(crt) {
			return fmt.Errorf("TLS enabled but keys not found in %s, run 'wwctl configure tls' to generate keys", path.Join(conf.Paths.Sysconfdir, "warewulf", "tls"))
		}
		httpsHandler := configureRootHandler(apiHandler)
		go func() {
			wwlog.Info("Starting HTTPS service on port %d", conf.Warewulf.TlsPort)
			if err := http.ListenAndServeTLS(":"+strconv.Itoa(conf.Warewulf.TlsPort), crt, key, httpsHandler); err != nil {
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
