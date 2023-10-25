package powercycle

import (
	"github.com/spf13/cobra"
)

var (
	powerCmd = &cobra.Command{
		DisableFlagsInUseLine: true,
		Use:                   "cycle [OPTIONS] [PATTERN ...]",
		Short:                 "Power cycle the given node(s)",
		Long:                  "This command cycles power for a set of nodes specified by PATTERN.",
		RunE:                  CobraRunE,
	}
)

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return powerCmd
}
