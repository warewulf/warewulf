package main

import (
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"github.com/pkg/errors"

	"github.com/hpcng/warewulf/internal/pkg/wwlog"
)

func NewDaemonConnection() (*DaemonConnection, error) {
	ret := DaemonConnection{}
	ret.init()
	err := ret.New()
	if err != nil {
		return &DaemonConnection{}, errors.Wrap(err, "Failed to prepare daemon connection")
	}
	return &ret, err
}

func runProductionEnv() error {
	err := os.Chdir("/")
	if err != nil {
		return errors.Wrap(err, "Failed to change dir")
	}
	wwlog.Println(wwlog.WARN, "Updating live file system LIVE, cancel now if this is in error")
	time.Sleep(5000 * time.Millisecond)
	return nil
}

func runTestEnv() error {
	wwlog.Printf(wwlog.WARN, "Called via: %s\n", os.Args[0])
	wwlog.Println(wwlog.WARN, "Runtime overlay is being put in '/warewulf/wwclient-test' rather than '/'")
	err := os.MkdirAll("/warewulf/wwclient-test", 0755)
	if err != nil {
		return errors.Wrap(err, "Failed to create dir")
	}

	err = os.Chdir("/warewulf/wwclient-test")
	if err != nil {
		return errors.Wrap(err, "Failed to change dir")
	}
	return nil
}

func main() {

	if os.Args[0] == "/warewulf/bin/wwclient" {
		err := runProductionEnv()
		if err != nil {
			wwlog.Printf(wwlog.ERROR, "Failed to run in production environment: %s\n", err)
			return
		}
	} else {
		err := runTestEnv()
		if err != nil {
			wwlog.Printf(wwlog.ERROR, "Failed to run in test environment: %s\n", err)
			return
		}
	}

	conn, err := NewDaemonConnection()
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "Failed to create daemon connection: %s\n", err)
		return
	}

	err = conn.AddHostname()
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "Failed to add hostname to query string: %s\n", err)
		return
	}

	err = conn.AddInterfaces()
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "Failed to add interfaces to query string: %s\n", err)
		return
	}

	conn.URL.RawQuery = conn.Values.Encode()
	wwlog.Printf(wwlog.INFO, "Encoded URL is %q\n", conn.URL.String())

	webclient := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				LocalAddr: &conn.TCPAddr,
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
	}

	// listen on SIGHUP
	sigs := make(chan os.Signal)
	signal.Notify(sigs, syscall.SIGHUP)

	go func() {
		for sig := range sigs {
			wwlog.Printf(wwlog.INFO, "Received SIGNAL: %s\n", sig)
			updateSystem(webclient, *conn)
		}
	}()

	for {
		updateSystem(webclient, *conn)

		if conn.updateInterval > 0 {
			time.Sleep(time.Duration(conn.updateInterval*1000) * time.Millisecond)
		} else {
			time.Sleep(30000 * time.Millisecond * 1000)
		}
	}
}

func updateSystem(webclient *http.Client, conn DaemonConnection) {
	var resp *http.Response
	counter := 0
	var err error

	for {
		resp, err = webclient.Get(conn.URL.String())
		if err == nil {
			break
		} else {
			if counter > 60 {
				counter = 0
			}
			if counter == 0 {
				log.Println(err)
			}
			counter++
		}
		time.Sleep(1000 * time.Millisecond)
	}

	if resp.StatusCode != 200 {
		wwlog.Printf(wwlog.WARN, "Not updating runtime overlay, got status code: %d\n", resp.StatusCode)
		time.Sleep(60000 * time.Millisecond)
		return
	}

	wwlog.Println(wwlog.INFO, "Updating system")
	command := exec.Command("/bin/sh", "-c", "gzip -dc | cpio -iu")
	command.Stdin = resp.Body
	err = command.Run()
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "Failed running CPIO: %s\n", err)
	}
}
