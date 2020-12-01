package delete

import "github.com/spf13/cobra"

var (
	baseCmd = &cobra.Command{
		Use:                "delete",
		Short:              "Delete",
		Long:               "Delete",
		RunE:				CobraRunE,
		Args: 				cobra.MinimumNArgs(1),
	}
	SetController string
)

func init() {

}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}
