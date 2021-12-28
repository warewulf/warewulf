package create

import (
	"github.com/spf13/cobra"
)

var (
	baseCmd = &cobra.Command{
		DisableFlagsInUseLine: true,
		Use:   "create [OPTIONS] {system|runtime} OVERLAY_NAME",
		Short: "Initialize a new Overlay",
		Long:  "This command creates a new empty system or runtime overlay named OVERLAY_NAME.",
		RunE:  CobraRunE,
		Args:  cobra.ExactArgs(2),
	}
)

func init() {
}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}
