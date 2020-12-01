package delete

import "github.com/spf13/cobra"

var (
	baseCmd = &cobra.Command{
		Use:                "delete",
		Short:              "Add a new node group",
		Long:               "Add a new node group ",
		RunE:				CobraRunE,
	}
	SetController string
)

func init() {
	baseCmd.PersistentFlags().StringVarP(&SetController, "controller", "c", "default", "Controller to add group to")
}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}
