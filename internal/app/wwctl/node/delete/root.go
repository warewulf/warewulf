package delete

import "github.com/spf13/cobra"

var (
	baseCmd = &cobra.Command{
		Use:     "delete [flags] [node pattern]...",
		Short:   "Delete a node from Warewulf",
		Long:    "This command will remove a node from the Warewulf node configuration.",
		Args:    cobra.MinimumNArgs(1),
		RunE:    CobraRunE,
		Aliases: []string{"rm", "del"},
	}
	SetYes            bool
	SetForce      string
	SetGroup      string
	SetController string
)

func init() {
	baseCmd.PersistentFlags().StringVarP(&SetForce, "force", "f", "", "Force node delete")
	baseCmd.PersistentFlags().StringVarP(&SetGroup, "group", "g", "default", "Set group to delete nodes from")
	baseCmd.PersistentFlags().StringVarP(&SetController, "controller", "c", "default", "Controller to add nodes to")
	baseCmd.PersistentFlags().BoolVarP(&SetYes, "yes", "y", false, "Set 'yes' to all questions asked")

}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}
