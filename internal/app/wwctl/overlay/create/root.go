package create

import (
	"github.com/spf13/cobra"
)

var (
	baseCmd = &cobra.Command{
		Use:                "create",
		Short:              "Initialize a new Overlay",
		Long:               "Create a new Warewulf provisioning overlay",
		RunE:				CobraRunE,
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
