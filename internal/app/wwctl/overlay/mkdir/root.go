package mkdir

import (
	"github.com/spf13/cobra"
)

var (
	baseCmd = &cobra.Command{
		DisableFlagsInUseLine: true,
		Use:   "mkdir [OPTIONS] {system|runtime} OVERLAY_NAME DIRECTORY",
		Short: "Create a new directory within an Overlay",
		Long:  "This command creates a new directory within the Warewulf OVERLAY_NAME.",
		RunE:  CobraRunE,
		Args:  cobra.MinimumNArgs(3),
	}
	PermMode int32
)

func init() {
	baseCmd.PersistentFlags().Int32VarP(&PermMode, "mode", "m", 0755, "Permission mode for directory")
}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}
