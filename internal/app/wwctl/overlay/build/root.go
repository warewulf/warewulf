package build

import "github.com/spf13/cobra"

var (
	baseCmd = &cobra.Command{
		DisableFlagsInUseLine: true,
		Use:   "build [OPTIONS] {system|runtime} OVERLAY_NAME",
		Short: "(Re)build an overlay",
		Long:  "This command builds a new system or runtime overlay named OVERLAY_NAME.",
		RunE:  CobraRunE,
		Args:  cobra.RangeArgs(0, 2),
	}
	BuildAll bool
)

func init() {
	baseCmd.PersistentFlags().BoolVarP(&BuildAll, "all", "a", false, "Build all overlays (runtime and system)")
}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}
