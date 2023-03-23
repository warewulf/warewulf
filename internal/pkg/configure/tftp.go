package configure

import (
	"fmt"
	"os"
	"path"

	"github.com/hpcng/warewulf/internal/pkg/util"
	"github.com/hpcng/warewulf/internal/pkg/warewulfconf"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
)

func TFTP() error {
	controller := warewulfconf.New()
	var tftpdir string = path.Join(controller.Paths.Tftpdir, "warewulf")

	err := os.MkdirAll(tftpdir, 0755)
	if err != nil {
		wwlog.Error("%s", err)
		return err
	}

	fmt.Printf("Writing PXE files to: %s\n", tftpdir)
	copyCheck := make(map[string]bool)
	for _, f := range controller.Tftp.IpxeBinaries {
		if copyCheck[f] {
			continue
		}
		copyCheck[f] = true
		err = util.SafeCopyFile(path.Join(controller.Paths.Datadir, f), path.Join(tftpdir, f))
		if err != nil {
			wwlog.Warn("ipxe binary could not be copied, booting may not work: %s", err)
		}
	}

	if !controller.Tftp.Enabled {
		wwlog.Info("Warewulf does not auto start TFTP services due to disable by warewulf.conf")
		os.Exit(0)
	}

	fmt.Printf("Enabling and restarting the TFTP services\n")
	err = util.SystemdStart(controller.Tftp.SystemdName)
	if err != nil {
		wwlog.Error("%s", err)
		return err
	}

	return nil
}
