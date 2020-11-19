package chmod

import "github.com/spf13/cobra"

var (
	baseCmd = &cobra.Command{
		Use:                "chmod [overlay name] [numeric mode] [file path]",
		Short:              "Change permissions within an overlay",
		Long:               "Change permissions within an overlay",
		RunE:				CobraRunE,
		Args: 				cobra.ExactArgs(3),

	}
	SystemOverlay bool
	NoOverlayUpdate bool
)

func init() {
	baseCmd.PersistentFlags().BoolVarP(&SystemOverlay, "system", "s", false, "Show System Overlays as well")
	baseCmd.PersistentFlags().BoolVarP(&NoOverlayUpdate, "noupdate", "n", false, "Don't update overlays")

}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}
