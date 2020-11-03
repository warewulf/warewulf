package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"time"
)




func main() {
	os.Chdir("/")

	// Setting up the connection manually so we can ensure a low port
	localAddr, err := net.ResolveIPAddr("ip", "localhost")
	if err != nil {
		panic(err)
	}

	localTCPAddr := net.TCPAddr{
		IP: localAddr.IP,
		Port: 987,
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

		for true {
			var err error
			fmt.Printf("Connecting ....\n")
			resp, err = webclient.Get("http://192.168.1.1:9873/runtime/xx-xx-xx-xx-xx")
			if err == nil {
				break
			} else {
				fmt.Println(err)
			}
			time.Sleep(1000 * time.Millisecond)
		}

		fmt.Printf("Connection accepted to remote host\n")
		command := exec.Command("cpio", "-i")
		stdin, err := command.StdinPipe()
		if err != nil {
			log.Fatal(err)
		}

		go func() {
			defer stdin.Close()
			io.Copy(stdin, resp.Body)
		}()

		command.Run()
		resp.Body.Close()

		time.Sleep(5000 * time.Millisecond)
	}
}
