package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"time"

	"github.com/hpcng/warewulf/internal/pkg/config"
)

func main() {
	if os.Args[0] == "/warewulf/bin/wwclient" {
		os.Chdir("/")
		log.Printf("Updating live file system LIVE, cancel now if this is in error")
		time.Sleep(5000 * time.Millisecond)
	} else {
		fmt.Printf("Called via: %s\n", os.Args[0])
		fmt.Printf("Runtime system-overlay is being put in '/warewulf/wwclient-test' rather than '/'\n")
		os.MkdirAll("/warewulf/wwclient-test", 0755)
		os.Chdir("/warewulf/wwclient-test")
	}

	config, err := config.New()
	if err != nil {
		fmt.Printf("ERROR: Could not load configuration file: %s\n", err)
		return
	}

	localTCPAddr := net.TCPAddr{}
	if config.InsecureRuntime == false {
		// Setup local port to something privileged (<1024)
		localTCPAddr.Port = 987
	} else {
		fmt.Printf("INFO: Running from an insecure port\n")
	}

	webclient := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				LocalAddr: &localTCPAddr,
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
				DualStack: true,
			}).DialContext,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
	}

	for true {
		var resp *http.Response
		counter := 0

		for true {
			var err error

			getString := fmt.Sprintf("http://%s:%d/overlay-runtime", config.Ipaddr, config.Port)
			resp, err = webclient.Get(getString)
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

//		defer resp.Body.Close()




		if resp.StatusCode != 200 {
			log.Printf("Not updating runtime system-overlay, got status code: %d\n", resp.StatusCode)
			time.Sleep(60000 * time.Millisecond)
			continue
		}


		log.Printf("Updating runtime system\n")
		command := exec.Command("/bin/cpio", "-iu")
		command.Stdin = resp.Body
		err := command.Run()
		if err != nil {
			log.Printf("ERROR: Failed running CPIO: %s\n", err)
		}

		time.Sleep(30000 * time.Millisecond)
	}
}
