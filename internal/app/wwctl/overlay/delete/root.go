package delete

import (
	"github.com/spf13/cobra"
)

var (
	baseCmd = &cobra.Command{
		Use:                "delete [overlay name]",
		Short:              "Delete Warewulf Overlay files",
		Long:               "Warewulf Delete overlay files",
		RunE:				CobraRunE,
		Args: 				cobra.MinimumNArgs(1),
		Aliases: 			[]string{"rm", "del"},
	}
	SystemOverlay bool
	Force bool
	Parents bool
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

