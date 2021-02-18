package build

import "github.com/spf13/cobra"

var (
	baseCmd = &cobra.Command{
		Use:   "build [flags] <overlay name>",
		Short: "(Re)build an overlay",
		Long:  "This command will build a system or runtime overlay.",
		RunE:  CobraRunE,
	}
	SystemOverlay bool
	BuildAll      bool
)

func init() {
	baseCmd.PersistentFlags().BoolVarP(&SystemOverlay, "system", "s", false, "Show System Overlays as well")
	baseCmd.PersistentFlags().BoolVarP(&BuildAll, "all", "a", false, "Build all overlays (runtime and system)")

}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}
