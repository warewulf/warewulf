package create

import (
	"github.com/spf13/cobra"
)

var (
	baseCmd = &cobra.Command{
		Use:   "create [flags] <overlay kind> <overlay name>",
		Short: "Initialize a new Overlay",
		Long:  "This command will create a new empty overlay.",
		RunE:  CobraRunE,
		Args:  cobra.ExactArgs(2),
	}
	NoOverlayUpdate bool
)

func init() {
	baseCmd.PersistentFlags().BoolVarP(&NoOverlayUpdate, "noupdate", "n", false, "Don't update overlays")
}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}
