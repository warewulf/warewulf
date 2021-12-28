package add

import "github.com/spf13/cobra"

var (
	baseCmd = &cobra.Command{
		DisableFlagsInUseLine: true,
		Use:   "add PROFILE",
		Short: "Add a new node profile",
		Long:  "This command adds a new named PROFILE.",
		RunE:  CobraRunE,
		Args:  cobra.MinimumNArgs(1),
	}
)

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}
