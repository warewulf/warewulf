package chown

import "github.com/spf13/cobra"

var (
	baseCmd = &cobra.Command{
		Use:   "chown [flags] <overlay> <path> <UID> [<GID>]",
		Short: "Change file ownership within an overlay",
		Long: "This command will allow you to change the ownership of a file within an\n" +
			"overlay.",
		RunE: CobraRunE,
		Args: cobra.RangeArgs(3, 4),
	}
	SystemOverlay   bool
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
