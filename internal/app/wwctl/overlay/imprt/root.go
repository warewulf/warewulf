package imprt

import "github.com/spf13/cobra"

var (
	baseCmd = &cobra.Command{
		DisableFlagsInUseLine: true,
		Use:                   "import [OPTIONS] OVERLAY_NAME FILE [NEW_NAME]",
		Short:                 "Import a file into a Warewulf Overlay",
		Long:                  "This command imports the FILE into the Warewulf OVERLAY_NAME.\nOptionally, the file can be renamed to NEW_NAME",
		RunE:                  CobraRunE,
		Args:                  cobra.RangeArgs(2, 3),
		Aliases:               []string{"cp"},
	}
	PermMode        int32
	NoOverlayUpdate bool
)

func init() {
	baseCmd.PersistentFlags().Int32VarP(&PermMode, "mode", "m", 0755, "Permission mode for directory")
	baseCmd.PersistentFlags().BoolVarP(&NoOverlayUpdate, "noupdate", "n", false, "Don't update overlays")
}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}
