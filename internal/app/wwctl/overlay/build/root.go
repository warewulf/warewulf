package build

import "github.com/spf13/cobra"

var (
	baseCmd = &cobra.Command{
		Use:                "build [overlay name]",
		Short:              "Build/Rebuild an overlay",
		Long:               "Build or rebuild a Warewulf overlay ",
		RunE:				CobraRunE,
	}
	SystemOverlay bool
	BuildAll bool
)

func init() {
	baseCmd.PersistentFlags().BoolVarP(&SystemOverlay, "system", "s", false, "Show System Overlays as well")
	baseCmd.PersistentFlags().BoolVarP(&BuildAll, "all", "a", false, "Build all overlays (runtime and system)")

}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}
