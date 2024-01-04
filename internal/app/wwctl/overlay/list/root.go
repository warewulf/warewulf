package list

import (
	"github.com/hpcng/warewulf/internal/pkg/overlay"
	"github.com/hpcng/warewulf/internal/pkg/util"
	"github.com/spf13/cobra"
)

type variables struct {
	listContents bool
	listLong     bool
	output       string
}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	vars := variables{}
	baseCmd := &cobra.Command{
		DisableFlagsInUseLine: true,
		Use:                   "list [OPTIONS] OVERLAY_NAME",
		Short:                 "List Warewulf Overlays and files",
		Long:                  "This command displays information about all Warewulf overlays or the specified\nOVERLAY_NAME. It also supports listing overlay content information.",
		RunE:                  CobraRunE(&vars),
		Args:                  cobra.MinimumNArgs(0),
		Aliases:               []string{"ls"},
		ValidArgs:             []string{"system", "runtime"},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) != 0 {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			list, _ := overlay.FindOverlays()
			return list, cobra.ShellCompDirectiveNoFileComp
		},
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return util.ValidOutput(vars.output)
		},
	}

	baseCmd.PersistentFlags().BoolVarP(&vars.listContents, "all", "a", false, "List the contents of overlays")
	baseCmd.PersistentFlags().BoolVarP(&vars.listLong, "long", "l", false, "List 'long' of all overlay contents")
	baseCmd.PersistentFlags().StringVarP(&vars.output, "output", "o", "text", "output format `json | text | yaml | csv`")
	return baseCmd
}
