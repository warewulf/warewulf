package tftp

import (
	"fmt"
	"os"
	"path"

	"github.com/hpcng/warewulf/internal/pkg/buildconfig"
	"github.com/hpcng/warewulf/internal/pkg/util"
	"github.com/hpcng/warewulf/internal/pkg/warewulfconf"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/spf13/cobra"
)

func CobraRunE(cmd *cobra.Command, args []string) error {
	return Configure(SetShow)
}

func Configure(show bool) error {
	controller, err := warewulfconf.New()
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "%s\n", err)
		os.Exit(1)
	}

	if buildconfig.TFTPDIR() == "" {
		wwlog.Printf(wwlog.ERROR, "Tftp root directory is not configured by build\n")
		os.Exit(1)
	}

	err = os.MkdirAll(path.Join(buildconfig.TFTPDIR(), "warewulf"), 0755)
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "%s\n", err)
		os.Exit(1)
	}

	if !show {

		fmt.Printf("Enabling and restarting the TFTP services\n")
		err = util.SystemdStart(controller.Tftp.SystemdName)
		if err != nil {
			wwlog.Printf(wwlog.ERROR, "%s\n", err)
			os.Exit(1)
		}
	}

	return nil
}
