package mkdir

import (
	"github.com/spf13/cobra"
	"github.com/warewulf/warewulf/internal/app/wwctl/completions"
)

var (
	baseCmd = &cobra.Command{
		DisableFlagsInUseLine: true,
		Use:                   "mkdir [OPTIONS] OVERLAY_NAME DIRECTORY",
		Short:                 "Create a new directory within an Overlay",
		Long:                  "This command creates a new directory within the Warewulf OVERLAY_NAME.",
		RunE:                  CobraRunE,
		Args:                  cobra.ExactArgs(2),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) == 0 {
				return completions.Overlays(cmd, args, toComplete)
			} else {
				return completions.None(cmd, args, toComplete)
			}
		},
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
