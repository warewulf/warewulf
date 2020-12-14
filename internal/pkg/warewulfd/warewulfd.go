package warewulfd

import (
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"net/http"
)

// TODO: https://github.com/danderson/netboot/blob/master/pixiecore/dhcp.go
// TODO: https://github.com/pin/tftp

func RunServer() error {

	wwlog.Printf(wwlog.DEBUG, "Registering handlers for the web service\n")

	http.HandleFunc("/ipxe/", IpxeSend)
	http.HandleFunc("/kernel/", KernelSend)
	http.HandleFunc("/kmods/", KmodsSend)
	http.HandleFunc("/container/", ContainerSend)
	http.HandleFunc("/overlay-system/", SystemOverlaySend)
	http.HandleFunc("/overlay-runtime", RuntimeOverlaySend)

	wwlog.Printf(wwlog.VERBOSE, "Starting HTTPD REST service\n")

	http.ListenAndServe(":9873", nil)

	return nil
}
