package warewulfd

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"strconv"

	"github.com/hpcng/warewulf/internal/pkg/warewulfconf"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
)

// TODO: https://github.com/danderson/netboot/blob/master/pixiecore/dhcp.go
// TODO: https://github.com/pin/tftp

func RunServer() error {

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP)

	go func() {
		for range c {
			err := LoadNodeDB()
			if err != nil {
				fmt.Printf("ERROR: Could not load database: %s\n", err)
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
	http.HandleFunc("/overlay-runtime", RuntimeOverlaySend)

	conf, err := warewulfconf.New()
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "Could not get Warewulf configuration: %s\n", err)
		os.Exit(1)
	}

	daemonPort := 9873
	if conf.Warewulf.Port != daemonPort {
		daemonPort = conf.Warewulf.Port
	} else {
		fmt.Printf("INFO: warewulfd port not configured, defaulting to 9873\n")
	}

	daemonLogf("Starting HTTPD REST service\n")

	err = http.ListenAndServe(":" + strconv.Itoa(daemonPort), nil)
	if err != nil {
		fmt.Printf("ERROR: Could not start listening service: %s\n", err)
		os.Exit(1)
	}

	return nil
}
