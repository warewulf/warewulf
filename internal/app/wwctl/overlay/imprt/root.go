package imprt

import "github.com/spf13/cobra"

var (
	baseCmd = &cobra.Command{
		Use:     "import [flags] (overlay kind) (overlay name) (host source) [overlay target]",
		Short:   "Import a file into a Warewulf Overlay",
		Long:    "This command will import a file into a given Warewulf overlay.",
		RunE:    CobraRunE,
		Args:    cobra.RangeArgs(3, 4),
		Aliases: []string{"cp"},
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
