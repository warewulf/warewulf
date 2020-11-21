package add

import "github.com/spf13/cobra"

var (
	baseCmd = &cobra.Command{
		Use:                "add",
		Short:              "Add new node",
		Long:               "Add new node ",
		RunE:				CobraRunE,
	}
)

func init() {
}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}
