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
		Args: 				cobra.ExactArgs(1),

	}
	SystemOverlay bool
	Force bool

)

func init() {
	baseCmd.PersistentFlags().BoolVarP(&SystemOverlay, "system", "s", false, "Show system overlays instead of runtime")
	baseCmd.PersistentFlags().BoolVar(&Force, "force", false, "Force deletion of a non-empty overlay")

}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}

