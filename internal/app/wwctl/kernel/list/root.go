package list

import "github.com/spf13/cobra"

var (
	baseCmd = &cobra.Command{
		DisableFlagsInUseLine: true,
		Use:                   "list [OPTIONS]",
		Short:                 "List available kernels",
		Long:                  "This command lists the kernels that are available in the imported containers.",
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
