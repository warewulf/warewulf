package list

import "github.com/spf13/cobra"

var (
	baseCmd = &cobra.Command{
		DisableFlagsInUseLine: true,
		Use:                   "list [OPTIONS]",
		Short:                 "List imported Kernel images",
		Long:                  "This command will list the kernels that have been imported into Warewulf.",
		RunE:                  CobraRunE,
		Args:                  cobra.ExactArgs(0),
		Aliases:               []string{"ls"},
	}
)

func init() {
}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}
