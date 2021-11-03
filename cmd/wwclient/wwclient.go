package main

import (
	"fmt"
	"github.com/pkg/errors"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"time"

	"github.com/hpcng/warewulf/internal/pkg/warewulfconf"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
)

type DaemonConnection struct {
	URL            url.URL
	TCPAddr        net.TCPAddr
	updateInterval int
}

func runProductionEnv() error {
	err := os.Chdir("/")
	if err != nil {
		return errors.Wrap(err, "failed to change dir")
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
		return errors.Wrap(err, "failed to create dir")
	}

	err = os.Chdir("/warewulf/wwclient-test")
	if err != nil {
		return errors.Wrap(err, "failed to change dir")
	}
	return nil
}

func prepDaemon() (DaemonConnection, error) {
	var ret DaemonConnection

	conf, err := warewulfconf.New()
	if err != nil {
		return DaemonConnection{}, errors.Wrap(err, "Could not get Warewulf configuration")
	}

	ret.updateInterval = conf.Warewulf.UpdateInterval

	if conf.Warewulf.Secure {
		ret.TCPAddr.Port = 987
	} else {
		wwlog.Println(wwlog.INFO, "Running from an insecure port")
	}

	// build the URL
	base := fmt.Sprintf("%s:%d", conf.Ipaddr, conf.Warewulf.Port)
	ret.URL = url.URL{Scheme: "http", Host: base}
	ret.URL.Path += "overlay-runtime"

	wwlog.Printf(wwlog.DEBUG, "baseURL: %s", ret.URL.String())

	return ret, nil
}

func InterfacesToValues() (url.Values, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return url.Values{}, errors.Wrap(err, "failed to obtain network interfaces")
	}
	params := url.Values{}
	for _, i := range interfaces {
		hwAddr := i.HardwareAddr.String()
		if len(hwAddr) == 0 {
			continue
		}
		params.Add("hwAddr", hwAddr)
	}
	return params, nil
}

func main() {

	if os.Args[0] == "/warewulf/bin/wwclient" {
		err := runProductionEnv()
		if err != nil {
			wwlog.Printf(wwlog.ERROR, "failed to run in production environment: %s\n", err)
			return
		}
	} else {
		err := runTestEnv()
		if err != nil {
			wwlog.Printf(wwlog.ERROR, "failed to run in test environment: %s\n", err)
			return
		}
	}

	conn, err := prepDaemon()
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "failed to prepare daemon connection: %s\n", err)
		return
	}
	params, err := InterfacesToValues()
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "failed to query Interfaces: %s\n", err)
		return
	}
	conn.URL.RawQuery = params.Encode()
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

	for {
		var resp *http.Response
		counter := 0

		for {
			var err error

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
			continue
		}

		wwlog.Println(wwlog.INFO, "Updating system")
		command := exec.Command("/bin/sh", "-c", "gzip -dc | cpio -iu")
		command.Stdin = resp.Body
		err := command.Run()
		if err != nil {
			wwlog.Printf(wwlog.ERROR, "ERROR: Failed running CPIO: %s\n", err)
		}

		if conn.updateInterval > 0 {
			time.Sleep(time.Duration(conn.updateInterval*1000) * time.Millisecond)
		} else {
			time.Sleep(30000 * time.Millisecond * 1000)
		}
	}
}
