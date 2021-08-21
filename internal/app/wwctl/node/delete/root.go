package delete

import "github.com/spf13/cobra"

var (
	baseCmd = &cobra.Command{
		Use:     "delete [flags] [exact node name]...",
		Short:   "Delete a node from Warewulf",
		Long:    "This command will remove a node from the Warewulf node configuration.",
		Args:    cobra.MinimumNArgs(1),
		RunE:    CobraRunE,
		Aliases: []string{"rm", "del"},
	}
	SetYes   bool
	SetForce string
)

func init() {
	baseCmd.PersistentFlags().StringVarP(&SetForce, "force", "f", "", "Force node delete")
	baseCmd.PersistentFlags().BoolVarP(&SetYes, "yes", "y", false, "Set 'yes' to all questions asked")

}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}
