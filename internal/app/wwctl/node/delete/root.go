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
)

func init() {
	baseCmd.PersistentFlags().StringVarP(&SetForce, "force", "f", "", "Force node delete")
	baseCmd.PersistentFlags().StringVarP(&SetGroup, "group", "g", "", "Set group to delete nodes from")

}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}
