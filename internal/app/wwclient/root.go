package wwclient

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path"
	"strings"
	"syscall"
	"time"

	"github.com/coreos/go-systemd/daemon"
	"github.com/google/uuid"
	"github.com/hpcng/warewulf/internal/pkg/buildconfig"
	"github.com/hpcng/warewulf/internal/pkg/pidfile"
	"github.com/hpcng/warewulf/internal/pkg/warewulfconf"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/spf13/cobra"
	"github.com/talos-systems/go-smbios/smbios"
)

var (
	rootCmd = &cobra.Command{
		Use:          "wwclient",
		Short:        "wwclient",
		Long:         "wwclient fetches the runtime overlay and puts it on the disk",
		RunE:         CobraRunE,
		SilenceUsage: true,
	}
	DebugFlag bool
	PIDFile   string
	Webclient *http.Client
)

func init() {
	rootCmd.PersistentFlags().BoolVarP(&DebugFlag, "debug", "d", false, "Run with debugging messages enabled.")
	rootCmd.PersistentFlags().StringVarP(&PIDFile, "pidfile", "p", "/var/run/wwclient.pid", "PIDFile to use")

}

// GetRootCommand returns the root cobra.Command for the application.
func GetRootCommand() *cobra.Command {
	// Run cobra
	return rootCmd
}

func CobraRunE(cmd *cobra.Command, args []string) error {
	conf, err := warewulfconf.New()
	if err != nil {
		return err
	}

	pid, err := pidfile.Write(PIDFile)
	if err != nil && pid == -1 {
		wwlog.Printf(wwlog.WARN, "%v. starting new wwclient", err)
	} else if err != nil && pid > 0 {
		return errors.New("found pidfile " + PIDFile + " not starting")
	}

	clientDir := buildconfig.WWCLIENTDIR()
	testDir := path.Join(clientDir, "wwclient-test")

	if os.Args[0] == path.Join(clientDir, "wwclient") {
		err := os.Chdir("/")
		if err != nil {
			wwlog.Printf(wwlog.ERROR, "failed to change dir: %s", err)
			_ = os.Remove(PIDFile)
			os.Exit(1)
		}
		log.Printf("Updating live file system LIVE, cancel now if this is in error")
		time.Sleep(5000 * time.Millisecond)
	} else {
		fmt.Printf("Called via: %s\n", os.Args[0])
		fmt.Printf("Runtime overlay is being put in '%s' rather than '/'\n", testDir)
		err := os.MkdirAll(testDir, 0755)
		if err != nil {
			wwlog.Printf(wwlog.ERROR, "failed to create dir: %s", err)
			_ = os.Remove(PIDFile)
			os.Exit(1)
		}

		err = os.Chdir(testDir)
		if err != nil {
			wwlog.Printf(wwlog.ERROR, "failed to change dir: %s", err)
			_ = os.Remove(PIDFile)
			os.Exit(1)
		}
	}

	localTCPAddr := net.TCPAddr{}
	if conf.Warewulf.Secure {
		// Setup local port to something privileged (<1024)
		localTCPAddr.Port = 987
		wwlog.Printf(wwlog.INFO, "Running from trusted port\n")
	}

	Webclient = &http.Client{
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
	tag := strings.ReplaceAll(x.AssetTagNumber(), " ", "_")

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

	duration := 300
	if conf.Warewulf.UpdateInterval > 0 {
		duration = conf.Warewulf.UpdateInterval
	}
	stopTimer := time.NewTimer(time.Duration(duration) * time.Second)
	// listen on SIGHUP
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGHUP, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		for {
			sig := <-sigs
			switch sig {
			case syscall.SIGHUP:
				log.Printf("Received SIGNAL: %s\n", sig)
				stopTimer.Stop()
				stopTimer.Reset(0)
			case syscall.SIGTERM, syscall.SIGINT:
				wwlog.Printf(wwlog.INFO, "termination wwclient!, %v", sig)
				cleanUp()
				os.Exit(0)
			}
		}
	}()
	var finishedInitialSync bool = false
	for {
		updateSystem(conf.Ipaddr, conf.Warewulf.Port, wwid, tag, localUUID)
		if !finishedInitialSync {
			// ignore error and status here, as this wouldn't change anything
			_, _ = daemon.SdNotify(false, daemon.SdNotifyReady)
			finishedInitialSync = true
		}

		<-stopTimer.C
		stopTimer.Reset(time.Duration(duration) * time.Second)
	}
}

func updateSystem(ipaddr string, port int, wwid string, tag string, localUUID uuid.UUID) {
	var resp *http.Response
	counter := 0
	for {
		var err error
		getString := fmt.Sprintf("http://%s:%d/overlay-runtime/%s?assetkey=%s&uuid=%s", ipaddr, port, wwid, tag, localUUID)
		wwlog.Printf(wwlog.DEBUG, "Making request: %s\n", getString)
		resp, err = Webclient.Get(getString)
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
		return
	}
	log.Printf("Updating system\n")
	command := exec.Command("/bin/sh", "-c", "gzip -dc | cpio -iu")
	command.Stdin = resp.Body
	err := command.Run()
	if err != nil {
		log.Printf("ERROR: Failed running CPIO: %s\n", err)
	}
}

func cleanUp() {
	err := pidfile.Remove(PIDFile)
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "could not remove pidfile: %s\n", err)
	}
}
