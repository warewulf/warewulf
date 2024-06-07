package configure

import (
	"os"

	"github.com/spf13/cobra"
	warewulfconf "github.com/warewulf/warewulf/internal/pkg/config"
	"github.com/warewulf/warewulf/internal/pkg/configure"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

func CobraRunE(cmd *cobra.Command, args []string) error {
	var err error
	if allFunctions {
		err = configure.DHCP()
		if err != nil {
			wwlog.Error("%s", err)
			os.Exit(1)
		}

		err = configure.NFS()
		if err != nil {
			wwlog.Error("%s", err)
			os.Exit(1)
		}

		err = configure.SSH(warewulfconf.Get().SSH.KeyTypes...)
		if err != nil {
			wwlog.Error("%s", err)
			os.Exit(1)
		}

		err = configure.TFTP()
		if err != nil {
			wwlog.Error("%s", err)
			os.Exit(1)
		}
		err = configure.Hostfile()
		if err != nil {
			wwlog.Error("%s", err)
			os.Exit(1)
		}

	} else {
		_ = cmd.Help()
		os.Exit(0)
	}

	return nil
}
