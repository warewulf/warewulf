package delete

import (
	"github.com/spf13/cobra"
)

var (
	baseCmd = &cobra.Command{
		Use:   "delete [flags] <overlay name> [overlay file]",
		Short: "Delete Warewulf Overlay or files",
		Long: "This command will delete files within an overlay or an entire overlay if no\n" +
			"files are given to remove (use with caution).",
		RunE:    CobraRunE,
		Args:    cobra.MinimumNArgs(1),
		Aliases: []string{"rm", "del"},
	}
	SystemOverlay   bool
	Force           bool
	Parents         bool
	NoOverlayUpdate bool
)

func init() {
	baseCmd.PersistentFlags().BoolVarP(&SystemOverlay, "system", "s", false, "Show system overlays instead of runtime")
	baseCmd.PersistentFlags().BoolVarP(&Force, "force", "f", false, "Force deletion of a non-empty overlay")
	baseCmd.PersistentFlags().BoolVarP(&Parents, "parents", "p", false, "Remove empty parent directories")
	baseCmd.PersistentFlags().BoolVarP(&NoOverlayUpdate, "noupdate", "n", false, "Don't update overlays")
}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}
