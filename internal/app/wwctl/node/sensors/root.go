package sensors

import (
	"github.com/spf13/cobra"
)

var (
	powerCmd = &cobra.Command{
		Use:   "sensors [flags] [node pattern]",
		Short: "Show node's IPMI sensor information",
		Long:  "Show IPMI sensors for a single node.",
		Args:  cobra.MinimumNArgs(1),
		RunE:  CobraRunE,
	}
	full bool
)

func init() {
	powerCmd.PersistentFlags().BoolVarP(&full, "full", "F", false, "show detailed output.")
}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return powerCmd
}
