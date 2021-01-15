package tftp

import (
	"fmt"
	"github.com/hpcng/warewulf/internal/pkg/staticfiles"
	"github.com/hpcng/warewulf/internal/pkg/util"
	"github.com/hpcng/warewulf/internal/pkg/warewulfconf"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/spf13/cobra"
	"os"
	"path"
)

func CobraRunE(cmd *cobra.Command, args []string) error {
	if SetShow == false && SetPersist == false {
		fmt.Println(cmd.Help())
		os.Exit(0)
	}

	return Configure(SetShow)
}

func Configure(show bool) error {
	controller, err := warewulfconf.New()
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "%s\n", err)
		os.Exit(1)
	}

	if controller.Tftp.TftpRoot == "" {
		wwlog.Printf(wwlog.ERROR, "Tftp root directory is not configured in warewulfd.conf\n")
		os.Exit(1)
	}

	if util.IsDir(controller.Tftp.TftpRoot) == false {
		wwlog.Printf(wwlog.ERROR, "Configured TFTP Root directory does not exist: %s\n", controller.Tftp.TftpRoot)
		os.Exit(1)
	}

	err = os.MkdirAll(path.Join(controller.Tftp.TftpRoot, "warewulf"), 0755)
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "%s\n", err)
		os.Exit(1)
	}

	if show == false {
		fmt.Printf("Writing PXE files to: %s\n", path.Join(controller.Tftp.TftpRoot, "warewulf"))
		err = staticfiles.WriteData("files/tftp/x86.efi", path.Join(controller.Tftp.TftpRoot, "warewulf/x86.efi"))
		if err != nil {
			wwlog.Printf(wwlog.ERROR, "%s\n", err)
			os.Exit(1)
		}
		err = staticfiles.WriteData("files/tftp/i386.efi", path.Join(controller.Tftp.TftpRoot, "warewulf/i386.efi"))
		if err != nil {
			wwlog.Printf(wwlog.ERROR, "%s\n", err)
			os.Exit(1)
		}
		err = staticfiles.WriteData("files/tftp/i386.kpxe", path.Join(controller.Tftp.TftpRoot, "warewulf/i386.kpxe"))
		if err != nil {
			wwlog.Printf(wwlog.ERROR, "%s\n", err)
			os.Exit(1)
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
