package chmod

import (
	"github.com/spf13/cobra"
	"github.com/warewulf/warewulf/internal/pkg/overlay"
)

var (
	baseCmd = &cobra.Command{
		DisableFlagsInUseLine: true,
		Use:                   "chmod [OPTIONS] OVERLAY_NAME FILENAME MODE",
		Short:                 "Change file permissions in an overlay",
		Long:                  "Changes the permissions of a single FILENAME within an overlay.\nYou can use any MODE format supported by the chmod command.",
		Example:               "wwctl overlay chmod default /etc/hostname.ww 0660",
		RunE:                  CobraRunE,
		Args:                  cobra.ExactArgs(3),
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
)

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}
