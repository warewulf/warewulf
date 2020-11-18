package copy

import "github.com/spf13/cobra"

var (
	baseCmd = &cobra.Command{
		Use:                "copy [overlay name] [source file] (dest location)",
		Short:              "Copy Warewulf Overlay files",
		Long:               "Warewulf Copy overlay files",
		RunE:				CobraRunE,
		Args: 				cobra.RangeArgs(2, 3),
		Aliases: 			[]string{"import"},
	}
	SystemOverlay bool
)

func init() {
	baseCmd.PersistentFlags().BoolVarP(&SystemOverlay, "system", "s", false, "Show system overlays instead of runtime")

}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}
