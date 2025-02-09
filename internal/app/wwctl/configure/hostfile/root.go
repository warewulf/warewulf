package hostfile

import (
	"github.com/spf13/cobra"
	"github.com/warewulf/warewulf/internal/app/wwctl/completions"
)

var (
	baseCmd = &cobra.Command{
		DisableFlagsInUseLine: true,
		Use:                   "hostfile  [OPTIONS]",
		Short:                 "update hostfile on master",
		Long:                  "Manage the hostfile on the master node\n",
		RunE:                  CobraRunE,
		Args:                  cobra.NoArgs,
		ValidArgsFunction:     completions.None,
	}
)

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}
