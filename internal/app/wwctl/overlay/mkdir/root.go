package mkdir

import (
	"github.com/spf13/cobra"
)

var (
	baseCmd = &cobra.Command{
		Use:   "mkdir [flags] <overlay name> <directory path>",
		Short: "Create a new directory within an Overlay",
		Long:  "This command will allow you to create a new file within a given Warewulf overlay.",
		RunE:  CobraRunE,
		Args:  cobra.MinimumNArgs(2),
	}
	SystemOverlay   bool
	PermMode        int32
	NoOverlayUpdate bool
)

func init() {
	baseCmd.PersistentFlags().BoolVarP(&SystemOverlay, "system", "s", false, "Show System Overlays as well")
	baseCmd.PersistentFlags().Int32VarP(&PermMode, "mode", "m", 0755, "Permission mode for directory")
	baseCmd.PersistentFlags().BoolVarP(&NoOverlayUpdate, "noupdate", "n", false, "Don't update overlays")
}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}
