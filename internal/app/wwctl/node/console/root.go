package console

import (
	"github.com/spf13/cobra"
	"github.com/warewulf/warewulf/internal/app/wwctl/completions"
)

var (
	powerCmd = &cobra.Command{
		DisableFlagsInUseLine: true,
		Use:                   "console [OPTIONS] NODENAME",
		Short:                 "Connect to IPMI console",
		Long:                  "Start a new IPMI console for NODENAME.",
		Args:                  cobra.MinimumNArgs(1),
		RunE:                  CobraRunE,
		ValidArgsFunction:     completions.Nodes(0), // no limit
	}
)

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return powerCmd
}
