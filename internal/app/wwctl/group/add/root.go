package add

import "github.com/spf13/cobra"

var (
	baseCmd = &cobra.Command{
		Use:   "add",
		Short: "Add a new node group",
		Long:  "Add a new node group ",
		RunE:  CobraRunE,
		Args:  cobra.MinimumNArgs(1),
	}
	SetController string
)

func init() {
	baseCmd.PersistentFlags().StringVarP(&SetController, "controller", "c", "localhost", "Controller to add group to")

}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}
