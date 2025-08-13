package wwclient

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"path"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/coreos/go-systemd/daemon"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"github.com/talos-systems/go-smbios/smbios"
	warewulfconf "github.com/warewulf/warewulf/internal/pkg/config"
	"github.com/warewulf/warewulf/internal/pkg/pidfile"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

var (
	rootCmd = &cobra.Command{
		Use:          "wwclient",
		Short:        "wwclient",
		Long:         "wwclient fetches the runtime overlay and puts it on the disk",
		RunE:         CobraRunE,
		SilenceUsage: true,
		Args:         cobra.NoArgs,
	}
	Once            bool
	DebugFlag       bool
	PIDFile         string
	Webclient       *http.Client
	WarewulfConfArg string
)

func init() {
	rootCmd.PersistentFlags().BoolVar(&Once, "once", false, "Run once and exit")
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
	if DebugFlag {
		wwlog.SetLogLevel(wwlog.DEBUG)
	} else {
		wwlog.SetLogLevel(wwlog.INFO)
	}

	conf := warewulfconf.Get()
	if WarewulfConfArg != "" {
		err = conf.Read(WarewulfConfArg, false)
	} else if os.Getenv("WAREWULFCONF") != "" {
		err = conf.Read(os.Getenv("WAREWULFCONF"), false)
	} else {
		err = conf.Read(warewulfconf.ConfigFile, false)
	}
	if err != nil {
		return
	}
	pid, err := pidfile.Write(PIDFile)
	if err != nil {
		if pid > 0 { // wwclient is already running
			return fmt.Errorf("%v: not starting", err)
		} else { // the pidfile is stale
			wwlog.Warn("%s: starting new wwclient", err)
		}
	}
	defer cleanUp()

	target := "/"
	if os.Args[0] == path.Join(conf.Paths.WWClientdir, "wwclient") {
		wwlog.Warn("updating live file system: cancel now if this is in error")
		time.Sleep(5000 * time.Millisecond)
	} else {
		target = "/warewulf/wwclient-test"

		fmt.Printf("Called via: %s\n", os.Args[0])
		fmt.Printf("Runtime overlay is being put in '%s' rather than '/'\n", target)
		fmt.Printf("For full functionality call with: %s\n", path.Join(conf.Paths.WWClientdir, "wwclient"))
		err := os.MkdirAll(target, 0755)
		if err != nil {
			return fmt.Errorf("failed to create dir: %w", err)
		}
	}

	localTCPAddr := net.TCPAddr{}
	if conf.WWClient != nil && conf.WWClient.Port > 0 {
		localTCPAddr.Port = int(conf.WWClient.Port)
		wwlog.Info("Running from configured port %d", conf.WWClient.Port)
	} else if conf.Warewulf.Secure() {
		// Setup local port to something privileged (<1024)
		localTCPAddr.Port = 987
		wwlog.Info("Running from trusted port: %d", localTCPAddr.Port)
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
			IdleConnTimeout:       2 * time.Duration(conf.Warewulf.UpdateInterval) * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
	}
	var localUUID uuid.UUID
	var tag string
	smbiosDump, smbiosErr := smbios.New()
	if smbiosErr == nil {
		sysinfoDump := smbiosDump.SystemInformation()
		localUUID, _ = sysinfoDump.UUID()
		x := smbiosDump.SystemEnclosure()
		tag = strings.ReplaceAll(x.AssetTagNumber(), " ", "_")
		if tag == "Unknown" {
			dmiOut, err := exec.Command("dmidecode", "-s", "chassis-asset-tag").Output()
			if err == nil {
				chassisAssetTag := strings.TrimSpace(string(dmiOut))
				if chassisAssetTag != "" {
					tag = chassisAssetTag
				}
			}
		}
	} else {
		// Raspberry Pi serial and DUID locations
		// /sys/firmware/devicetree/base/serial-number
		// /sys/firmware/devicetree/base/chosen/rpi-duid
		piSerial, err := os.ReadFile("/sys/firmware/devicetree/base/serial-number")
		if err != nil {
			return fmt.Errorf("could not get SMBIOS info: %w", smbiosErr)
		}
		localUUID = uuid.NewSHA1(uuid.NameSpaceURL, []byte("http://raspberrypi.com/serial-number/"+string(piSerial)))
		tag = "Unknown"
	}

	wwlog.Debug("uuid: %s", localUUID.String())
	wwlog.Debug("assetkey: %s", tag)

	cmdline, err := os.ReadFile("/proc/cmdline")
	if err != nil {
		return fmt.Errorf("could not read from /proc/cmdline: %w", err)
	}
	wwid, err := parseWWIDFromCmdline(string(cmdline))
	if err != nil {
		return fmt.Errorf("failed to parse wwid: %w", err)
	}

	// Dereference wwid from [interface] for cases that cannot have /proc/cmdline set by bootloader
	if string(wwid[0]) == "[" {
		iface := wwid[1 : len(wwid)-1]
		wwid_tmp, err := os.ReadFile(fmt.Sprintf("/sys/class/net/%s/address", iface))
		if err != nil {
			return fmt.Errorf("'wwid' cannot be dereferenced from /sys/class/net: %w", err)
		}
		wwid = strings.TrimSuffix(string(wwid_tmp), "\n")
		wwlog.Info("Dereferencing wwid from [%s] to %s", iface, wwid)
	}

	wwlog.Debug("wwid: %s", wwid)

	duration := 300
	if conf.Warewulf.UpdateInterval > 0 {
		duration = conf.Warewulf.UpdateInterval
	}
	stopTimer := time.NewTimer(time.Duration(duration) * time.Second)
	// listen on SIGHUP
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGHUP, syscall.SIGTERM, syscall.SIGINT)

	// Add a channel to signal main loop to exit gracefully
	exitChan := make(chan bool, 1)

	go func() {
		for {
			sig := <-sigs
			switch sig {
			case syscall.SIGHUP:
				wwlog.Info("received signal: %s", sig)
				stopTimer.Stop()
				stopTimer.Reset(0)
			case syscall.SIGTERM, syscall.SIGINT:
				wwlog.Info("terminating wwclient, %v", sig)
				// Signal main loop to exit instead of calling os.Exit(0)
				exitChan <- true
				return
			}
		}
	}()
	var finishedInitialSync bool = false
	ipaddr := os.Getenv("WW_IPADDR")
	if ipaddr == "" {
		ipaddr = conf.Ipaddr
	}

	for {
		updateSystem(target, ipaddr, conf.Warewulf.Port, wwid, tag, localUUID)
		if !finishedInitialSync {
			// Notify systemd that the service has started successfully.
			//
			// Ignoring error and status, as this wouldn't change anything.
			_, _ = daemon.SdNotify(false, daemon.SdNotifyReady)
			finishedInitialSync = true
		}

		if Once {
			return nil
		}

		// Check for exit signal or timer
		select {
		case <-exitChan:
			wwlog.Info("gracefully shutting down")
			return nil
		case <-stopTimer.C:
			stopTimer.Reset(time.Duration(duration) * time.Second)
		}
	}
}

// parseWWIDFromCmdline extracts the wwid parameter from kernel command line
func parseWWIDFromCmdline(cmdline string) (string, error) {
	params := strings.Fields(cmdline)

	for _, param := range params {
		if strings.HasPrefix(param, "wwid=") {
			wwid := strings.TrimPrefix(param, "wwid=")
			if wwid == "" {
				return "", fmt.Errorf("wwid parameter is empty")
			}
			return wwid, nil
		}
	}

	return "", fmt.Errorf("wwid parameter not found in kernel command line")
}

func updateSystem(target string, ipaddr string, port int, wwid string, tag string, localUUID uuid.UUID) {
	var resp *http.Response
	counter := 0
	for {
		var err error
		values := &url.Values{}
		values.Set("assetkey", tag)
		values.Set("uuid", localUUID.String())
		values.Set("stage", "runtime")
		values.Set("compress", "gz")
		getURL := &url.URL{
			Scheme:   "http",
			Host:     fmt.Sprintf("%s:%d", ipaddr, port),
			Path:     fmt.Sprintf("provision/%s", wwid),
			RawQuery: values.Encode(),
		}
		wwlog.Debug("making request: %s", getURL)
		resp, err = Webclient.Get(getURL.String())
		if err == nil {
			break
		} else {
			if counter > 60 {
				counter = 0
			}
			if counter == 0 {
				wwlog.Error("%s", err)
			}
			counter++
		}
		time.Sleep(1000 * time.Millisecond)
	}
	if resp.StatusCode != 200 {
		wwlog.Warn("not applying runtime overlay: got status code: %d", resp.StatusCode)
		time.Sleep(60000 * time.Millisecond)
		return
	}

	wwlog.Info("applying runtime overlay")

	// unpack overlay into a temporary directory
	tempDir, err := os.MkdirTemp("", "wwclient-")
	if err != nil {
		wwlog.Error("failed to create temp directory: %s", err)
		return
	}
	defer os.RemoveAll(tempDir)
	wwlog.Debug("unpacking runtime overlay to %s", tempDir)
	command := exec.Command("/bin/sh", "-c", fmt.Sprintf("gzip -dc | cpio -iu --directory=%s", tempDir))
	command.Stdin = resp.Body
	err = command.Run()
	if err != nil {
		wwlog.Error("failed running cpio: %s", err)
		return
	}

	// Atomically move files from temp directory to current working directory
	err = atomicApplyOverlay(tempDir, target)
	if err != nil {
		wwlog.Error("failed to apply overlay: %s", err)
	}
}

func atomicApplyOverlay(srcDir, destDir string) error {
	return filepath.Walk(srcDir, func(srcPath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Calculate relative path from srcDir
		relPath, err := filepath.Rel(srcDir, srcPath)
		if err != nil {
			return err
		}

		// Skip the root directory itself
		if relPath == "." {
			return nil
		}

		destPath := filepath.Join(destDir, relPath)

		if info.IsDir() {
			// Create directory if it doesn't exist
			wwlog.Debug("Ensuring directory exists: %s", destPath)
			err := os.MkdirAll(destPath, info.Mode())
			if err != nil {
				return err
			}
			// Ensure permissions are updated even if directory already existed
			wwlog.Debug("Updating permissions for directory: %s", destPath)
			err = os.Chmod(destPath, info.Mode())
			if err != nil {
				return fmt.Errorf("failed to update directory permissions for %s: %w", destPath, err)
			}
			// Update ownership and timestamps
			if stat, ok := info.Sys().(*syscall.Stat_t); ok {
				wwlog.Debug("Updating ownership for directory: %s", destPath)
				err := os.Chown(destPath, int(stat.Uid), int(stat.Gid))
				if err != nil {
					wwlog.Warn("failed to update ownership for directory %s: %s", destPath, err)
				}
			}
			err = os.Chtimes(destPath, info.ModTime(), info.ModTime())
			if err != nil {
				wwlog.Warn("failed to update timestamps for directory %s: %s", destPath, err)
			}
			return nil

		} else if info.Mode()&os.ModeSymlink != 0 {
			// Handle symbolic links
			linkTarget, err := os.Readlink(srcPath)
			if err != nil {
				return fmt.Errorf("failed to read symlink %s: %w", srcPath, err)
			}

			// Create a temporary symlink in same directory as the destination.
			// This ensures the temp file is on the same filesystem as the final
			// destination, so the rename placement is atomic.
			destParent := filepath.Dir(destPath)

			// Ensure destination directory exists
			if err := os.MkdirAll(destParent, 0755); err != nil {
				return fmt.Errorf("failed to create destination directory %s: %w", destParent, err)
			}

			tempFile, err := os.CreateTemp(destParent, ".wwclient-tmp-")
			if err != nil {
				return fmt.Errorf("failed to create temp file for symlink %s: %w", destPath, err)
			}
			tempPath := tempFile.Name()
			_ = tempFile.Close()
			os.Remove(tempPath) // Remove the regular file so we can create a symlink

			wwlog.Debug("Creating temporary symlink: %s -> %s", tempPath, linkTarget)
			err = os.Symlink(linkTarget, tempPath)
			if err != nil {
				return fmt.Errorf("failed to create temporary symlink %s: %w", tempPath, err)
			}

			// Update ownership if possible (note: lchown for symlinks)
			if stat, ok := info.Sys().(*syscall.Stat_t); ok {
				wwlog.Debug("Updating ownership for temporary symlink: %s", tempPath)
				err := syscall.Lchown(tempPath, int(stat.Uid), int(stat.Gid))
				if err != nil {
					wwlog.Warn("failed to update ownership for temporary symlink %s: %s", tempPath, err)
				}
			}

			// Atomic rename - this will be atomic since both files are in the same directory
			wwlog.Debug("Moving symlink %s to %s", tempPath, destPath)
			err = os.Rename(tempPath, destPath)
			if err != nil {
				os.Remove(tempPath)
				return fmt.Errorf("failed to atomically move symlink %s to %s: %w", tempPath, destPath, err)
			}

			return nil

		} else {
			// Create a temporary file in same directory as the destination.
			// This ensures the temp file is on the same filesystem as the final
			// destination, so the rename placement is atomic.
			destParent := filepath.Dir(destPath)

			// Ensure destination directory exists
			if err := os.MkdirAll(destParent, 0755); err != nil {
				return fmt.Errorf("failed to create destination directory %s: %w", destParent, err)
			}

			tempFile, err := os.CreateTemp(destParent, ".wwclient-tmp-")
			if err != nil {
				return fmt.Errorf("failed to create temp file for %s: %w", destPath, err)
			}
			tempPath := tempFile.Name()
			_ = tempFile.Close()

			// Copy file content and metadata
			wwlog.Debug("Copying file from %s to temp location %s", srcPath, tempPath)
			err = copyFile(srcPath, tempPath, info)
			if err != nil {
				os.Remove(tempPath)
				return fmt.Errorf("failed to copy %s to temp location: %w", srcPath, err)
			}

			// Atomic rename - this will be atomic since both files are in the same directory
			// (and thus on the same filesystem)
			wwlog.Debug("Moving %s to %s", tempPath, destPath)
			err = os.Rename(tempPath, destPath)
			if err != nil {
				os.Remove(tempPath)
				return fmt.Errorf("failed to atomically move %s to %s: %w", tempPath, destPath, err)
			}
		}

		return nil
	})
}

func copyFile(src, dst string, srcInfo os.FileInfo) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, srcInfo.Mode())
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return err
	}

	// Preserve timestamps and ownership if possible
	if stat, ok := srcInfo.Sys().(*syscall.Stat_t); ok {
		wwlog.Debug("Updating ownership for file: %s", dst)
		err = os.Chown(dst, int(stat.Uid), int(stat.Gid))
		if err != nil {
			wwlog.Warn("failed to update ownership for file %s: %s", dst, err)
		}
	}
	err = os.Chtimes(dst, srcInfo.ModTime(), srcInfo.ModTime())
	if err != nil {
		wwlog.Warn("failed to update timestamps for file %s: %s", dst, err)
	}

	return nil
}

func cleanUp() {
	err := pidfile.Remove(PIDFile)
	if err != nil {
		wwlog.Error("could not remove pidfile: %s", err)
	}
}
