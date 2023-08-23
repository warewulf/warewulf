package wwclient

import (
	"errors"
	"fmt"
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
	warewulfconf "github.com/hpcng/warewulf/internal/pkg/config"
	"github.com/hpcng/warewulf/internal/pkg/pidfile"
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
	DebugFlag       bool
	PIDFile         string
	Webclient       *http.Client
	WarewulfConfArg string
)

func init() {
	rootCmd.PersistentFlags().BoolVarP(&DebugFlag, "debug", "d", false, "Run with debugging messages enabled.")
	rootCmd.PersistentFlags().StringVarP(&PIDFile, "pidfile", "p", "/var/run/wwclient.pid", "PIDFile to use")
	rootCmd.PersistentFlags().StringVar(&WarewulfConfArg, "warewulfconf", "", "Set the warewulf configuration file")

}

// GetRootCommand returns the root cobra.Command for the application.
func GetRootCommand() *cobra.Command {
	// Run cobra
	return rootCmd
}

func CobraRunE(cmd *cobra.Command, args []string) (err error) {
	conf := warewulfconf.Get()
	if WarewulfConfArg != "" {
		err = conf.Read(WarewulfConfArg)
	} else if os.Getenv("WAREWULFCONF") != "" {
		err = conf.Read(os.Getenv("WAREWULFCONF"))
	} else {
		err = conf.Read(warewulfconf.ConfigFile)
	}
	if err != nil {
		return
	}
	pid, err := pidfile.Write(PIDFile)
	if err != nil && pid == -1 {
		wwlog.Warn("%v. starting new wwclient", err)
	} else if err != nil && pid > 0 {
		return errors.New("found pidfile " + PIDFile + " not starting")
	}

	if os.Args[0] == path.Join(conf.Paths.WWClientdir, "wwclient") {
		err := os.Chdir("/")
		if err != nil {
			wwlog.Error("failed to change dir: %s", err)
			_ = os.Remove(PIDFile)
			os.Exit(1)
		}
		log.Printf("Updating live file system LIVE, cancel now if this is in error")
		time.Sleep(5000 * time.Millisecond)
	} else {
		fmt.Printf("Called via: %s\n", os.Args[0])
		fmt.Printf("Runtime overlay is being put in '/warewulf/wwclient-test' rather than '/'\n")
		fmt.Printf("For full functionality call with: %s\n", path.Join(conf.Paths.WWClientdir, "wwclient"))
		err := os.MkdirAll("/warewulf/wwclient-test", 0755)
		if err != nil {
			wwlog.Error("failed to create dir: %s", err)
			_ = os.Remove(PIDFile)
			os.Exit(1)
		}

		err = os.Chdir("/warewulf/wwclient-test")
		if err != nil {
			wwlog.Error("failed to change dir: %s", err)
			_ = os.Remove(PIDFile)
			os.Exit(1)
		}
	}

	localTCPAddr := net.TCPAddr{}
	if conf.Warewulf.Secure {
		// Setup local port to something privileged (<1024)
		localTCPAddr.Port = 987
		wwlog.Info("Running from trusted port")
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
		wwlog.Error("Could not get SMBIOS info: %s", err)
		os.Exit(1)
	}
	sysinfoDump := smbiosDump.SystemInformation()
	localUUID, _ := sysinfoDump.UUID()
	x := smbiosDump.SystemEnclosure()
	tag := strings.ReplaceAll(x.AssetTagNumber(), " ", "_")

	cmdline, err := os.ReadFile("/proc/cmdline")
	if err != nil {
		wwlog.Error("Could not read from /proc/cmdline: %s", err)
		os.Exit(1)
	}

	wwid_tmp := strings.Split(string(cmdline), "wwid=")
	if len(wwid_tmp) < 2 {
		wwlog.Error("'wwid' is not defined in /proc/cmdline")
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
				wwlog.Info("termination wwclient!, %v", sig)
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
		getString := fmt.Sprintf("http://%s:%d/provision/%s?assetkey=%s&uuid=%s&stage=runtime&compress=gz", ipaddr, port, wwid, tag, localUUID)
		wwlog.Debug("Making request: %s", getString)
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
		wwlog.Error("could not remove pidfile: %s", err)
	}
}
