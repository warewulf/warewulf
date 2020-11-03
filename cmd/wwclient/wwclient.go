package main

import (
	"bufio"
	"fmt"
	"net"
	"net/http"
	"time"
)




func main() {
	localAddr, err := net.ResolveIPAddr("ip", "localhost")
	if err != nil {
		panic(err)
	}

	// You also need to do this to make it work and not give you a
	// "mismatched local address type ip"
	// This will make the ResolveIPAddr a TCPAddr without needing to
	// say what SRC port number to use.
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
			resp, err = webclient.Get("http://localhost:9873/files/runtime/xx-xx-xx-xx-xx-xx")
			if err == nil {
				break
			} else {
				fmt.Println(err)
			}
			time.Sleep(1000 * time.Millisecond)
		}

		fmt.Println("Response status:", resp.Status)
		scanner := bufio.NewScanner(resp.Body)
		for i := 0; scanner.Scan() && i < 5; i++ {
			fmt.Println(scanner.Text())
		}
		if err := scanner.Err(); err != nil {
			panic(err)
		}
		resp.Body.Close()
		time.Sleep(5000 * time.Millisecond)
	}
}
