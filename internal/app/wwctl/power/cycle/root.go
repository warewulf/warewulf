package powercycle

import (
	"github.com/spf13/cobra"
)

var (
	powerCmd = &cobra.Command{
		Use:   "cycle [flags] (node pattern)...",
		Short: "Power cycle the given node(s)",
		Long:  "This command will cycle the power for a given set of nodes.",
		RunE:  CobraRunE,
	}
)

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return powerCmd
}
