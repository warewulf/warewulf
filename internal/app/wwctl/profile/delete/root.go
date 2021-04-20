package delete

import "github.com/spf13/cobra"

var (
	baseCmd = &cobra.Command{
		Use:   "delete [flags] <profile pattern>",
		Short: "Delete a node profile",
		Long:  "This command will delete a node profile.",
		RunE:  CobraRunE,
		Args:  cobra.MinimumNArgs(1),
	}
	SetYes bool
)

func init() {
	baseCmd.PersistentFlags().BoolVarP(&SetYes, "yes", "y", false, "Set 'yes' to all questions asked")
}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}
