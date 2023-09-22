package mkdir

import (
	"github.com/hpcng/warewulf/internal/pkg/overlay"
	"github.com/spf13/cobra"
)

var (
	baseCmd = &cobra.Command{
		DisableFlagsInUseLine: true,
		Use:                   "mkdir [OPTIONS] OVERLAY_NAME DIRECTORY",
		Short:                 "Create a new directory within an Overlay",
		Long:                  "This command creates a new directory within the Warewulf OVERLAY_NAME.",
		RunE:                  CobraRunE,
		Args:                  cobra.MinimumNArgs(2),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) == 0 {
				list, _ := overlay.FindOverlays()
				return list, cobra.ShellCompDirectiveNoFileComp
			} else if len(args) == 1 {
				ret, err := overlay.OverlayGetFiles(args[0])
				if err == nil {
					return ret, cobra.ShellCompDirectiveNoFileComp
				}
			}
			return []string{""}, cobra.ShellCompDirectiveNoFileComp
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
