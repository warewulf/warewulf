package configure

import (
	"os"
	"path"
	"strings"

	"github.com/opencontainers/selinux/go-selinux"
	warewulfconf "github.com/warewulf/warewulf/internal/pkg/config"
	"github.com/warewulf/warewulf/internal/pkg/util"
	"github.com/warewulf/warewulf/internal/pkg/warewulfd"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
	"golang.org/x/sys/unix"
)

func TFTP() (err error) {
	controller := warewulfconf.Get()
	var tftpdir string = path.Join(controller.TFTP.TftpRoot, "warewulf")
	oldMask := unix.Umask(0)
	defer unix.Umask(oldMask)
	err = os.MkdirAll(tftpdir, 0755)
	if err != nil {
		return
	}

	if controller.Warewulf.GrubBoot() {
		err := warewulfd.CopyShimGrub()
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

	// Set SELinux context if configured
	if selinux.GetEnabled() && getSelinuxContext(controller.TFTP.Selinux) != "" {
		currentLabel, err := selinux.FileLabel(tftpdir)
		if err != nil {
			wwlog.Warn("Failed to get current SELinux context for %s: %s", tftpdir, err)
		} else if currentLabel != controller.TFTP.Selinux {
			wwlog.Info("Setting SELinux context for %s to %s", tftpdir, controller.TFTP.Selinux)
			if err := selinux.Chcon(tftpdir, controller.TFTP.Selinux, true); err != nil {
				wwlog.Warn("Failed to set SELinux context on %s: %s", tftpdir, err)
			} else {
				wwlog.Info("To make the SELinux policy permanent, run: semanage fcontext -a '%s' '%s(/.*)?'", controller.TFTP.Selinux, tftpdir)
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

func getSelinuxContext(value string) string {
	if strings.ToLower(value) == "default" {
		return "system_u:object_r:public_content_t:s0"
	} else {
		return value
	}
}
