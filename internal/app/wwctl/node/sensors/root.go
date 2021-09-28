package sensors

import (
	"github.com/hpcng/warewulf/internal/pkg/node"
	"github.com/spf13/cobra"
)

var (
	powerCmd = &cobra.Command{
		Use:   "sensors [flags] [node pattern]",
		Short: "Show node's IPMI sensor information",
		Long:  "Show IPMI sensors for a single node.",
		Args:  cobra.MinimumNArgs(1),
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
	full bool
)

func init() {
	powerCmd.PersistentFlags().BoolVarP(&full, "full", "F", false, "show detailed output.")
}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return powerCmd
}
