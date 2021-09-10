package warewulfd

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
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

	daemonLogf("Starting HTTPD REST service\n")

	err = http.ListenAndServe(":9873", nil)
	if err != nil {
		fmt.Printf("ERROR: Could not start listening service: %s\n", err)
		os.Exit(1)
	}

	return nil
}
