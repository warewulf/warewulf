package reset

import (
	"github.com/spf13/cobra"
	"github.com/warewulf/warewulf/internal/pkg/node"
)

var (
	powerCmd = &cobra.Command{
		DisableFlagsInUseLine: true,
		Use:                   "reset [OPTIONS] [PATTERN ...]",
		Short:                 "Issue a reset to node(s)",
		Long:                  "This command will issue a reset to a set of nodes specified by PATTERN.",
		RunE:                  CobraRunE,
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) != 0 {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}

			nodeDB, _ := node.New()
			nodes, _ := nodeDB.FindAllNodes()
			var node_names []string
			for _, node := range nodes {
				node_names = append(node_names, node.Id())
			}
			return node_names, cobra.ShellCompDirectiveNoFileComp
		},
	}
)

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return powerCmd
}
