package console

import (
	"github.com/spf13/cobra"
)

var (
	powerCmd = &cobra.Command{
		Use:   "console [flags] [node pattern]",
		Short: "Connect to IPMI console",
		Long:  "Start IPMI console for a singe node.",
		Args:  cobra.MinimumNArgs(1),
		RunE:  CobraRunE,
	}
)

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return powerCmd
}
