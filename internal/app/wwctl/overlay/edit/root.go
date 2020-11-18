package edit

import (
	"github.com/spf13/cobra"
)

var (
	baseCmd = &cobra.Command{
		Use:                "edit [overlay name] [file path]",
		Short:              "Edit Warewulf Overlay files",
		Long:               "Warewulf edit overlay files",
		RunE:				CobraRunE,
		Args: 				cobra.ExactArgs(2),

	}
	SystemOverlay bool
	ListFiles bool
	CreateDirs bool
)

func init() {
	baseCmd.PersistentFlags().BoolVarP(&SystemOverlay, "system", "s", false, "Show system overlays instead of runtime")
	baseCmd.PersistentFlags().BoolVarP(&ListFiles, "files", "f", false, "List files contained within a given overlay")
	baseCmd.PersistentFlags().BoolVarP(&CreateDirs, "parents", "p", false, "Create any necessary parent directories")

}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}


