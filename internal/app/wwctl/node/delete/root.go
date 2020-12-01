package delete

import "github.com/spf13/cobra"

var (
	baseCmd = &cobra.Command{
		Use:                "delete",
		Short:              "Set node configurations",
		Long:               "Set node configurations ",
		RunE:				CobraRunE,
	}
	SetForce string
	SetGroup string
	SetController string
)

func init() {
	baseCmd.PersistentFlags().StringVarP(&SetForce, "force", "f", "", "Force node delete")
	baseCmd.PersistentFlags().StringVarP(&SetGroup, "group", "g", "default", "Set group to delete nodes from")
	baseCmd.PersistentFlags().StringVarP(&SetController, "controller", "c", "default", "Controller to add nodes to")

}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}
