package edit

import (
	"github.com/spf13/cobra"
)

var (
	baseCmd = &cobra.Command{
		Use:   "edit [flags] <overlay name> <file path>",
		Short: "Edit/Create a file within a Warewulf Overlay",
		Long: "This command will allow you to edit or create a new file within a given\n" +
			"overlay. Note: when creating files ending in a '.ww' suffix this will always be\n" +
			"parsed as a Warewulf template file, and the suffix will be removed automatically.",
		RunE: CobraRunE,
		Args: cobra.ExactArgs(2),
	}
	SystemOverlay   bool
	ListFiles       bool
	CreateDirs      bool
	PermMode        int32
	NoOverlayUpdate bool
)

func init() {
	baseCmd.PersistentFlags().BoolVarP(&SystemOverlay, "system", "s", false, "Show system overlays instead of runtime")
	baseCmd.PersistentFlags().BoolVarP(&ListFiles, "files", "f", false, "List files contained within a given overlay")
	baseCmd.PersistentFlags().BoolVarP(&CreateDirs, "parents", "p", false, "Create any necessary parent directories")
	baseCmd.PersistentFlags().Int32VarP(&PermMode, "mode", "m", 0755, "Permission mode for directory")
	baseCmd.PersistentFlags().BoolVarP(&NoOverlayUpdate, "noupdate", "n", false, "Don't update overlays")
}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}
