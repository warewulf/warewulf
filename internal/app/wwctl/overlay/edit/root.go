package edit

import (
	"github.com/spf13/cobra"
	"github.com/warewulf/warewulf/internal/app/wwctl/completions"
)

var (
	baseCmd = &cobra.Command{
		DisableFlagsInUseLine: true,
		Use:                   "edit [OPTIONS] OVERLAY_NAME FILE",
		Short:                 "Edit or create a file within a Warewulf Overlay",
		Long:                  "This command will open the FILE for editing or create a new file within the\nOVERLAY_NAME. Note: files created with a '.ww' suffix will always be\nparsed as Warewulf template files, and the suffix will be removed automatically.",
		RunE:                  CobraRunE,
		Args:                  cobra.ExactArgs(2),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) < 2 {
				return completions.OverlayAndFiles(cmd, args, toComplete)
			}
			return completions.None(cmd, args, toComplete)
		},
	}
	CreateDirs bool
	PermMode   int32
)

func init() {
	baseCmd.PersistentFlags().BoolVarP(&CreateDirs, "parents", "p", false, "Create any necessary parent directories")
	baseCmd.PersistentFlags().Int32VarP(&PermMode, "mode", "m", 0755, "Permission mode for directory")
}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}
