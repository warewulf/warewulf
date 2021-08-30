package powerreset

import (
	"github.com/spf13/cobra"
)

var (
	powerCmd = &cobra.Command{
		Use:   "reset",
		Short: "Issue a reset to the given node(s)",
		Long:  "This command will issue a reset to the given set of nodes.",
		RunE:  CobraRunE,
	}
)

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return powerCmd
}
