package list

import (
	"github.com/spf13/cobra"
	"github.com/warewulf/warewulf/internal/pkg/overlay"
)

var (
	baseCmd = &cobra.Command{
		DisableFlagsInUseLine: true,
		Use:                   "list [OPTIONS] OVERLAY_NAME",
		Short:                 "List Warewulf Overlays and files",
		Long:                  "This command displays information about all Warewulf overlays or the specified\nOVERLAY_NAME. It also supports listing overlay content information.",
		RunE:                  CobraRunE,
		Args:                  cobra.MinimumNArgs(0),
		Aliases:               []string{"ls"},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) != 0 {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			list := overlay.FindOverlays()
			return list, cobra.ShellCompDirectiveNoFileComp
		},
	}
	ListContents bool
	ListLong     bool
)

func init() {
	baseCmd.PersistentFlags().BoolVarP(&ListContents, "all", "a", false, "List the contents of overlays")
	baseCmd.PersistentFlags().BoolVarP(&ListLong, "long", "l", false, "List 'long' of all overlay contents")

}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}
