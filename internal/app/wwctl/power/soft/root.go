package powersoft

import (
	"github.com/spf13/cobra"
)

var (
	powerCmd = &cobra.Command{
		Use:   "soft",
		Short: "Gracefully shuts down the given node(s)",
		Long:  "This command will gracefully shutdown the given set of nodes.",
		RunE:  CobraRunE,
	}
)

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return powerCmd
}
