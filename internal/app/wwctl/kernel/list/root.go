package list

import "github.com/spf13/cobra"

var (
	baseCmd = &cobra.Command{
		Use:                "list",
		Short:              "List Installed Kernel Images",
		Long:               "List installed kernel images",
		RunE:				CobraRunE,
		Args: 				cobra.ExactArgs(0),
	}
)

func init() {
}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}