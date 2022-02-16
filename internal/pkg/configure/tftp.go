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

func configureTFTP() error {
	controller, err := warewulfconf.New()
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "%s\n", err)
		return err
	}

	err = os.MkdirAll(tftpdir, 0755)
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "%s\n", err)
		return err
	}

	fmt.Printf("Writing PXE files to: %s\n", tftpdir)
	for _, f := range [4]string{"x86.efi", "i386.efi", "i386.kpxe", "arm64.efi"} {
		err = util.CopyFileSafe(path.Join(buildconfig.DATADIR(), "warewulf", "ipxe", f), path.Join(tftpdir, f))
		if err != nil {
			wwlog.Printf(wwlog.ERROR, "%s\n", err)
		}
		return err
	}

	fmt.Printf("Enabling and restarting the TFTP services\n")
	err = util.SystemdStart(controller.Tftp.SystemdName)
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "%s\n", err)
		return err
	}

	return nil
}
