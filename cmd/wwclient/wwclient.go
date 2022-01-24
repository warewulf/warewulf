package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/hpcng/warewulf/internal/pkg/warewulfconf"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/talos-systems/go-smbios/smbios"
)

func main() {
	if os.Args[0] == "/warewulf/bin/wwclient" {
		err := os.Chdir("/")
		if err != nil {
			wwlog.Printf(wwlog.ERROR, "failed to change dir: %s", err)
			os.Exit(1)
		}
		log.Printf("Updating live file system LIVE, cancel now if this is in error")
		time.Sleep(5000 * time.Millisecond)
	} else {
		fmt.Printf("Called via: %s\n", os.Args[0])
		fmt.Printf("Runtime overlay is being put in '/warewulf/wwclient-test' rather than '/'\n")
		err := os.MkdirAll("/warewulf/wwclient-test", 0755)
		if err != nil {
			wwlog.Printf(wwlog.ERROR, "failed to create dir: %s", err)
			os.Exit(1)
		}

		err = os.Chdir("/warewulf/wwclient-test")
		if err != nil {
			wwlog.Printf(wwlog.ERROR, "failed to change dir: %s", err)
			os.Exit(1)
		}
	}

	conf, err := warewulfconf.New()
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "Could not get Warewulf configuration: %s\n", err)
		os.Exit(1)
	}

	localTCPAddr := net.TCPAddr{}
	if conf.Warewulf.Secure {
		// Setup local port to something privileged (<1024)
		localTCPAddr.Port = 987
		wwlog.Printf(wwlog.INFO, "Running from trusted port\n")
	}

	webclient := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				LocalAddr: &localTCPAddr,
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
	}

	smbiosDump, err := smbios.New()
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "Could not get SMBIOS info: %s\n", err)
		os.Exit(1)
	}
	sysinfoDump := smbiosDump.SystemInformation()
	localUUID, _ := sysinfoDump.UUID()
	x := smbiosDump.SystemEnclosure()
	tag := x.AssetTagNumber()

	cmdline, err := ioutil.ReadFile("/proc/cmdline")
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "Could not read from /proc/cmdline: %s\n", err)
		os.Exit(1)
	}

	wwid_tmp := strings.Split(string(cmdline), "wwid=")
	if len(wwid_tmp) < 2 {
		wwlog.Printf(wwlog.ERROR, "'wwid' is not defined in /proc/cmdline\n")
		os.Exit(1)
	}

	wwid := strings.Split(wwid_tmp[1], " ")[0]

	for {
		var resp *http.Response
		counter := 0

		for {
			var err error

			getString := fmt.Sprintf("http://%s:%d/overlay-runtime/%s?assetkey=%s&uuid=%s", conf.Ipaddr, conf.Warewulf.Port, wwid, tag, localUUID)
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

		if resp.StatusCode != 200 {
			log.Printf("Not updating runtime overlay, got status code: %d\n", resp.StatusCode)
			time.Sleep(60000 * time.Millisecond)
			continue
		}

		log.Printf("Updating system\n")
		command := exec.Command("/bin/sh", "-c", "gzip -dc | cpio -iu")
		command.Stdin = resp.Body
		err := command.Run()
		if err != nil {
			log.Printf("ERROR: Failed running CPIO: %s\n", err)
		}

		if conf.Warewulf.UpdateInterval > 0 {
			time.Sleep(time.Duration(conf.Warewulf.UpdateInterval*1000) * time.Millisecond)
		} else {
			time.Sleep(30000 * time.Millisecond * 1000)
		}
	}
}
