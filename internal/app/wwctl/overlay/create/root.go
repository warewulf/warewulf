package create

import (
	"github.com/spf13/cobra"
)

var (
	baseCmd = &cobra.Command{
		DisableFlagsInUseLine: true,
		Use:                   "create [OPTIONS] OVERLAY_NAME",
		Short:                 "Initialize a new Overlay",
		Aliases:               []string{"new", "add"},
		Long:                  "This command creates a new empty overlay with the given OVERLAY_NAME.",
		RunE:                  CobraRunE,
		Args:                  cobra.ExactArgs(1),
	}
)

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}
