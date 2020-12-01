package delete

import "github.com/spf13/cobra"

var (
	baseCmd = &cobra.Command{
		Use:                "delete",
		Short:              "Delete profiles",
		Long:               "Profile configurations ",
		RunE:				CobraRunE,
	}
)

func init() {
}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}
