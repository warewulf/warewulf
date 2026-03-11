package configure

import (
	"fmt"
	"os"
	"path"

	warewulfconf "github.com/warewulf/warewulf/internal/pkg/config"
	"github.com/warewulf/warewulf/internal/pkg/image"
	"github.com/warewulf/warewulf/internal/pkg/util"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
	"golang.org/x/sys/unix"
)

func TFTP() (err error) {
	controller := warewulfconf.Get()
	oldMask := unix.Umask(0)
	defer unix.Umask(oldMask)

	// Check if TftpRoot exists, create and restore context if needed
	if _, err := os.Stat(controller.TFTP.TftpRoot); err != nil {
		err = os.MkdirAll(controller.TFTP.TftpRoot, 0755)
		if err != nil {
			return err
		}
		if err := util.RestoreSELinuxContext(controller.TFTP.TftpRoot); err != nil {
			wwlog.Warn("failed to restore SELinux context for %s: %s", controller.TFTP.TftpRoot, err)
		}
	}

	// Create tftpdir if needed
	var tftpdir string = path.Join(controller.TFTP.TftpRoot, "warewulf")
	err = os.MkdirAll(tftpdir, 0755)
	if err != nil {
		return
	}

	if controller.Warewulf.GrubBoot() {
		err := CopyShimGrub()
		if err != nil {
			wwlog.Warn("error when copying shim/grub binaries: %s", err)
		}
	} else {
		wwlog.Info("Writing PXE files to: %s", tftpdir)
		copyCheck := make(map[string]bool)
		for _, f := range controller.TFTP.IpxeBinaries {
			if !path.IsAbs(f) {
				f = path.Join(controller.Paths.Ipxesource, f)
			}
			if copyCheck[f] {
				continue
			}
			copyCheck[f] = true
			err = util.SafeCopyFile(f, path.Join(tftpdir, path.Base(f)))
			if err != nil {
				wwlog.Warn("ipxe binary could not be copied, booting may not work: %s", err)
			}
		}
	}

	if !controller.TFTP.Enabled() {
		wwlog.Warn("Warewulf does not auto start TFTP services due to disable by warewulf.conf")
		return nil
	}

	wwlog.Info("Enabling and restarting the TFTP services")
	err = util.SystemdStart(controller.TFTP.SystemdName)
	if err != nil {
		return
	}

	return nil
}

func CopyShimGrub() (err error) {
	conf := warewulfconf.Get()
	wwlog.Debug("copy shim and grub binaries from host")
	shimPath := image.ShimFind("")
	if shimPath == "" {
		return fmt.Errorf("no shim found on the host os")
	}
	err = util.CopyFile(shimPath, path.Join(conf.TFTP.TftpRoot, "warewulf", "shim.efi"))
	if err != nil {
		return err
	}
	_ = os.Chmod(path.Join(conf.TFTP.TftpRoot, "warewulf", "shim.efi"), 0o755)
	grubPath := image.GrubFind("")
	if grubPath == "" {
		return fmt.Errorf("no grub found on host os")
	}
	err = util.CopyFile(grubPath, path.Join(conf.TFTP.TftpRoot, "warewulf", "grub.efi"))
	if err != nil {
		return err
	}
	_ = os.Chmod(path.Join(conf.TFTP.TftpRoot, "warewulf", "grub.efi"), 0o755)
	err = util.CopyFile(grubPath, path.Join(conf.TFTP.TftpRoot, "warewulf", "grubx64.efi"))
	_ = os.Chmod(path.Join(conf.TFTP.TftpRoot, "warewulf", "grubx64.efi"), 0o755)

	return
}
