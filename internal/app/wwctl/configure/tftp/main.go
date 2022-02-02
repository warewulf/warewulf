package tftp

import (
	"fmt"
	"os"
	"path"

	"github.com/hpcng/warewulf/internal/pkg/staticfiles"
	"github.com/hpcng/warewulf/internal/pkg/util"
	"github.com/hpcng/warewulf/internal/pkg/warewulfconf"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/spf13/cobra"
)

var tftpdir string = warewulfconf.Config("tftproot")

func CobraRunE(cmd *cobra.Command, args []string) error {
	return Configure(SetShow)
}

func Configure(show bool) error {
	controller, err := warewulfconf.New()
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "%s\n", err)
		os.Exit(1)
	}

	err = os.MkdirAll(tftpdir, 0755)
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "%s\n", err)
		os.Exit(1)
	}

	if !show {
		fmt.Printf("Writing PXE files to: %s\n", tftpdir)
		for _, f := range [4]string{"x86.efi", "i386.efi", "i386.kpxe", "arm64.efi"} {
			err = staticfiles.WriteData(path.Join("files/tftp", f), path.Join(tftpdir, f))
			if err != nil {
				wwlog.Printf(wwlog.ERROR, "%s\n", err)
				os.Exit(1)
			}
		}

		fmt.Printf("Enabling and restarting the TFTP services\n")
		err = util.SystemdStart(controller.Tftp.SystemdName)
		if err != nil {
			wwlog.Printf(wwlog.ERROR, "%s\n", err)
			os.Exit(1)
		}
	}

	return nil
}
