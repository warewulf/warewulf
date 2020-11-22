package add

import "github.com/spf13/cobra"

var (
	baseCmd = &cobra.Command{
		Use:                "add",
		Short:              "Add new node",
		Long:               "Add new node ",
		RunE:				CobraRunE,
	}
	SetGroup string
)

func init() {
	baseCmd.PersistentFlags().StringVarP(&SetGroup, "group", "g", "default", "Set group to add nodes to")
}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}
