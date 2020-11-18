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
	RmEmptyDirs bool
)

func init() {
	baseCmd.PersistentFlags().BoolVarP(&SystemOverlay, "system", "s", false, "Show system overlays instead of runtime")
	baseCmd.PersistentFlags().BoolVarP(&Force, "force", "f", false, "Force deletion of a non-empty overlay")
	baseCmd.PersistentFlags().BoolVarP(&RmEmptyDirs, "empty", "e", false, "Remove empty directories")

}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}

