package build

import (
	"github.com/spf13/cobra"
)

var (
	baseCmd = &cobra.Command{
		Use:   "build (kernel version | node search pattern)",
		Short: "Kernel Image Build",
		Long:  "Build kernel images",
		RunE:  CobraRunE,
		Args:  cobra.MinimumNArgs(1),
	}
	BuildAll bool
	ByNode   bool
)

func init() {
	baseCmd.PersistentFlags().BoolVarP(&BuildAll, "all", "a", false, "Build all overlays (runtime and system)")
	baseCmd.PersistentFlags().BoolVarP(&ByNode, "node", "n", false, "Build overlay for a particular node(s)")
}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}
