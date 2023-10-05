package configure

import (
	"os"
	"path"

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
	err = os.MkdirAll(tftpdir, 0755)
	if err != nil {
		return
	}
	_ = unix.Umask(oldMask)

	if controller.Warewulf.GrubBoot {
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
	if !controller.TFTP.Enabled {
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
