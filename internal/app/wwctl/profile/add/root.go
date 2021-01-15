package add

import "github.com/spf13/cobra"

var (
	baseCmd = &cobra.Command{
		Use:   "add [flags] <profile name>",
		Short: "Add a new node profile",
		Long:  "This command will add a new node profile.",
		RunE:  CobraRunE,
		Args:  cobra.MinimumNArgs(1),
	}
)

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}
