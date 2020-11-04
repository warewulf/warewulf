package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"

	//	"os/exec"
	"time"
)

func main() {
	if os.Args[0] == "/warewulf/bin/wwclient" {
		os.Chdir("/")
		log.Printf("Updating live file system LIVE, cancel now if this is in error")
		time.Sleep(5000 * time.Millisecond)
	} else {
		fmt.Printf("Called via: %s\n", os.Args[0])
		fmt.Printf("Runtime overlay is being put in '/warewulf/wwclient-test' rather than '/'\n")
		os.MkdirAll("/warewulf/wwclient-test", 0755)
		os.Chdir("/warewulf/wwclient-test")
	}

	// Setup local port to 987
	localTCPAddr := net.TCPAddr{
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
		counter := 0

		for true {
			var err error

			resp, err = webclient.Get("http://192.168.1.1:9873/overlay-runtime")
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
			log.Printf("Not updating runtime overlay, got status code: %d\n", resp.StatusCode)
			time.Sleep(60000 * time.Millisecond)
			continue
		}

/*
		// TODO: Turn all of this into a pipe instead of having to use a tmpfile which
		//       I tried to get working, but when running on a node, it always gave a
		//       trying to write on closed file descriptor... This maybe ugly, but it
		//       works.
		tmpfile := fmt.Sprintf("/tmp/.wwclient-%s", util.RandomString(14))
		tmpFD, _ := os.Create(tmpfile)
		defer tmpFD.Close()
		io.Copy(tmpFD, resp.Body)
		tmpFD.Close()

		err := exec.Command("cpio", "-i", "-F", tmpfile).Run()
		if err != nil {
			fmt.Printf("%s", err)
		}

		os.Remove(tmpfile)

*/
		log.Printf("Updating runtime system\n")
		command := exec.Command("/bin/cpio", "-i")
		command.Stdin = resp.Body
		err := command.Run()
		if err != nil {
			log.Printf("ERROR: Failed running CPIO: %s\n", err)
		}
		/*
				command.Wait()
				stdin, err := command.StdinPipe()
				if err != nil {
					log.Println(err)
				}
				defer stdin.Close()

				go func() {
					bytes, err := io.Copy(stdin, resp.Body)
					if err != nil {
						log.Printf("ERROR: io.Copy() failed: %s\n", err)
					} else {
						log.Printf("Updated the runtime overlay (recv: %d)\n", bytes)
					}

				}()
				command.Run()
		*/
//		defer webclient.CloseIdleConnections()


		time.Sleep(30000 * time.Millisecond)
	}
}
