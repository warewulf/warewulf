package show

import (
	"log"

	"github.com/hpcng/warewulf/internal/pkg/node"
	"github.com/hpcng/warewulf/internal/pkg/overlay"
	"github.com/spf13/cobra"
)

var (
	baseCmd = &cobra.Command{
		DisableFlagsInUseLine: true,
		Use:                   "show [OPTIONS] OVERLAY_NAME FILE",
		Short:                 "Show (cat) a file within a Warewulf Overlay",
		Long:                  "This command displays the contents of FILE within OVERLAY_NAME.",
		RunE:                  CobraRunE,
		Aliases:               []string{"cat"},
		Args:                  cobra.ExactArgs(2),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) != 0 {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			list, _ := overlay.FindOverlays()
			return list, cobra.ShellCompDirectiveNoFileComp
		},
	}
	NodeName string
	Quiet    bool
)

func init() {
	baseCmd.PersistentFlags().StringVarP(&NodeName, "render", "r", "", "node used for the variables in the template")
	if err := baseCmd.RegisterFlagCompletionFunc("render", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		nodeDB, _ := node.ReadNodeYaml()
		nodes, _ := nodeDB.GetAllNodeInfo()
		var node_names []string
		for _, node := range nodes {
			node_names = append(node_names, node.Id.Get())
		}
		return node_names, cobra.ShellCompDirectiveNoFileComp
	}); err != nil {
		log.Println(err)
	}
	baseCmd.PersistentFlags().BoolVarP(&Quiet, "quiet", "q", false, "do not print information if multiple, backup files are written")
}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}
