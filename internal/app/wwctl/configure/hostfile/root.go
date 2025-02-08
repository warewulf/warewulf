package hostfile

import "github.com/spf13/cobra"

var (
	baseCmd = &cobra.Command{
		DisableFlagsInUseLine: true,
		Use:                   "hostfile  [OPTIONS]",
		Short:                 "update hostfile on master",
		Long:                  "Manage the hostfile on the master node\n",
		RunE:                  CobraRunE,
		Args:                  cobra.ExactArgs(0),
	}
)

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}
