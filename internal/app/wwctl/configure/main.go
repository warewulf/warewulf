package configure

import (
	"os"

	"github.com/hpcng/warewulf/internal/pkg/configure"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/spf13/cobra"
)

func CobraRunE(cmd *cobra.Command, args []string) error {
	var err error
	if allFunctions {
		err = configure.Dhcp()
		if err != nil {
			wwlog.Error("%s\n", err)
			os.Exit(1)
		}

		err = configure.NFS()
		if err != nil {
			wwlog.Error("%s\n", err)
			os.Exit(1)
		}

		err = configure.SSH()
		if err != nil {
			wwlog.Error("%s\n", err)
			os.Exit(1)
		}

		err = configure.TFTP()
		if err != nil {
			wwlog.Error("%s\n", err)
			os.Exit(1)
		}
	} else {
		_ = cmd.Help()
		os.Exit(0)
	}

	return nil
}
