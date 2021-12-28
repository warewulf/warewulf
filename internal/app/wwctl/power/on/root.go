package poweron

import (
	"github.com/hpcng/warewulf/internal/pkg/node"
	"github.com/spf13/cobra"
)

var (
	powerCmd = &cobra.Command{
		Use:   "on [OPTIONS] [PATTERN ...]",
		Short: "Power on the given node(s)",
		Long:  "This command will power on a set of nodes specified by PATTERN.",
		RunE:  CobraRunE,
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) != 0 {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}

			nodeDB, _ := node.New()
			nodes, _ := nodeDB.FindAllNodes()
			var node_names []string
			for _, node := range nodes {
				node_names = append(node_names, node.Id.Get())
			}
			return node_names, cobra.ShellCompDirectiveNoFileComp
		},
	}
)

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return powerCmd
}
