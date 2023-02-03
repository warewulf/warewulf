package configure

import (
	"fmt"
	"os"
	"path"

	"github.com/hpcng/warewulf/internal/pkg/buildconfig"
	"github.com/hpcng/warewulf/internal/pkg/util"
	"github.com/hpcng/warewulf/internal/pkg/warewulfconf"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
)

var tftpdir string = path.Join(buildconfig.TFTPDIR(), "warewulf")

func TFTP() error {
	controller, err := warewulfconf.New()
	if err != nil {
		wwlog.Error("%s", err)
		return err
	}

	err = os.MkdirAll(tftpdir, 0755)
	if err != nil {
		wwlog.Error("%s", err)
		return err
	}

	fmt.Printf("Writing PXE files to: %s\n", tftpdir)
	for _, f := range [4]string{"x86_64.efi", "x86_64.kpxe", "arm64.efi"} {
		err = util.SafeCopyFile(path.Join(buildconfig.DATADIR(), "warewulf", "ipxe", f), path.Join(tftpdir, f))
		if err != nil {
			wwlog.Error("%s", err)
			return err
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
