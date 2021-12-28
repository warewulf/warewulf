package edit

import (
	"github.com/spf13/cobra"
)

var (
	baseCmd = &cobra.Command{
		DisableFlagsInUseLine: true,
		Use:   "edit [OPTIONS] {system|runtime} OVERLAY_NAME FILE",
		Short: "Edit or create a file within a Warewulf Overlay",
		Long: "This command will open the FILE for editing or create a new file within the\n" +
			"OVERLAY_NAME. Note: files created with a '.ww' suffix will always be\n" +
			"parsed as Warewulf template files, and the suffix will be removed automatically.",
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
