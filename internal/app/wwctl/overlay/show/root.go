package show

import (
	"github.com/spf13/cobra"
)

var (
	baseCmd = &cobra.Command{
		Use:                "show",
		Short:              "Show Warewulf Overlay objects",
		Long:               "Warewulf show overlay objects",
		RunE:				CobraRunE,
		Aliases: 			[]string{"cat"},
		Args: 				cobra.ExactArgs(2),
	}
	SystemOverlay bool
)

func init() {
	baseCmd.PersistentFlags().BoolVarP(&SystemOverlay, "system", "s", false, "Show System Overlays as well")
}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}