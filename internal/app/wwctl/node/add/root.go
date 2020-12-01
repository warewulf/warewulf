package add

import "github.com/spf13/cobra"

var (
	baseCmd = &cobra.Command{
		Use:                "add",
		Short:              "Add new node",
		Long:               "Add new node ",
		RunE:				CobraRunE,
		Args: 				cobra.MinimumNArgs(1),
	}
	SetGroup string
	SetController string
)

func init() {
	baseCmd.PersistentFlags().StringVarP(&SetGroup, "group", "g", "default", "Group to add nodes to")
	baseCmd.PersistentFlags().StringVarP(&SetController, "controller", "c", "default", "Controller to add nodes to")
}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}
