package imprt

import (
	"github.com/hpcng/warewulf/internal/pkg/overlay"
	"github.com/spf13/cobra"
)

var (
	baseCmd = &cobra.Command{
		DisableFlagsInUseLine: true,
		Use:                   "import [OPTIONS] OVERLAY_NAME FILE [NEW_NAME]",
		Short:                 "Import a file into a Warewulf Overlay",
		Long:                  "This command imports the FILE into the Warewulf OVERLAY_NAME.\nOptionally, the file can be renamed to NEW_NAME",
		RunE:                  CobraRunE,
		Args:                  cobra.RangeArgs(2, 3),
		Aliases:               []string{"cp"},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) != 0 {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			list, _ := overlay.FindOverlays()
			return list, cobra.ShellCompDirectiveNoFileComp
		},
	}
	NoOverlayUpdate bool
	CreateDirs      bool
)

func init() {
	baseCmd.PersistentFlags().BoolVarP(&NoOverlayUpdate, "noupdate", "n", false, "Don't update overlays")
	baseCmd.PersistentFlags().BoolVarP(&CreateDirs, "parents", "p", false, "Create any necessary parent directories")
}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}
