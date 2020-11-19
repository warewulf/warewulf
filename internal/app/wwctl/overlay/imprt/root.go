package imprt

import "github.com/spf13/cobra"

var (
	baseCmd = &cobra.Command{
		Use:                "import [overlay name] [source file] (dest location)",
		Short:              "Import Warewulf Overlay files",
		Long:               "Warewulf Import overlay files",
		RunE:				CobraRunE,
		Args: 				cobra.RangeArgs(2, 3),
		Aliases: 			[]string{"cp"},
	}
	SystemOverlay bool
	PermMode int32
	NoOverlayUpdate bool
)

func init() {
	baseCmd.PersistentFlags().BoolVarP(&SystemOverlay, "system", "s", false, "Show system overlays instead of runtime")
	baseCmd.PersistentFlags().Int32VarP(&PermMode, "mode", "m", 0755, "Permission mode for directory")
	baseCmd.PersistentFlags().BoolVarP(&NoOverlayUpdate, "noupdate", "n", false, "Don't update overlays")
}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}
