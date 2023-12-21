package configure

import (
	"fmt"
	"os"
	"path"

	warewulfconf "github.com/hpcng/warewulf/internal/pkg/config"
	"github.com/hpcng/warewulf/internal/pkg/util"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
)

func TFTP() error {
	controller := warewulfconf.Get()
	var tftpdir string = path.Join(controller.Paths.Tftpdir, "warewulf")

	err := os.MkdirAll(tftpdir, 0755)
	if err != nil {
		wwlog.Error("%s", err)
		return err
	}

	fmt.Printf("Writing PXE files to: %s\n", tftpdir)
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

	if !controller.TFTP.Enabled {
		wwlog.Info("Warewulf does not auto start TFTP services due to disable by warewulf.conf")
		os.Exit(0)
	}

	fmt.Printf("Enabling and restarting the TFTP services\n")
	err = util.SystemdStart(controller.TFTP.SystemdName)
	if err != nil {
		wwlog.Error("%s", err)
		return err
	}

	return nil
}
