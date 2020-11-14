package overlay

import (
	"github.com/hpcng/warewulf/internal/app/wwctl/overlay/list"
	"github.com/hpcng/warewulf/internal/app/wwctl/overlay/show"

	"github.com/spf13/cobra"
)

var (
	overlayCmd = &cobra.Command{
		Use:                "overlay",
		Short:              "Warewulf Overlay Management",
		Long:               "Management interface for Warewulf overlays",
	}
	test bool
)

func init() {
	overlayCmd.PersistentFlags().BoolVarP(&test, "test", "t", false, "Testing.")

	overlayCmd.AddCommand(list.GetCommand())
	overlayCmd.AddCommand(show.GetCommand())

}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return overlayCmd
}
