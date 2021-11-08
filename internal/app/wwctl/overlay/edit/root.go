package edit

import (
	"github.com/spf13/cobra"
)

var (
	baseCmd = &cobra.Command{
		Use:   "edit [flags] (overlay kind) (overlay name) (file path)",
		Short: "Edit/Create a file within a Warewulf Overlay",
		Long: "This command will allow you to edit or create a new file within a given\n" +
			"overlay. Note: when creating files ending in a '.ww' suffix this will always be\n" +
			"parsed as a Warewulf template file, and the suffix will be removed automatically.",
		RunE: CobraRunE,
		Args: cobra.ExactArgs(3),
	}
	ListFiles  bool
	CreateDirs bool
	PermMode   int32
)

func init() {
	baseCmd.PersistentFlags().BoolVarP(&ListFiles, "files", "f", false, "List files contained within a given overlay")
	baseCmd.PersistentFlags().BoolVarP(&CreateDirs, "parents", "p", false, "Create any necessary parent directories")
	baseCmd.PersistentFlags().Int32VarP(&PermMode, "mode", "m", 0755, "Permission mode for directory")
}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}
