package off

import (
	"github.com/spf13/cobra"
	"github.com/warewulf/warewulf/internal/pkg/node"
)

type variables struct {
	Showcmd bool
	Fanout  int
}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	vars := variables{}
	powerCmd := &cobra.Command{
		DisableFlagsInUseLine: true,
		Use:                   "off [OPTIONS] [PATTERN ...]",
		Short:                 "Power off the given node(s)",
		Long:                  "This command will shutdown power to a set of nodes specified by PATTERN.",
		RunE:                  CobraRunE(&vars),
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
	powerCmd.PersistentFlags().BoolVarP(&vars.Showcmd, "show", "s", false, "only show command which will be executed")
	powerCmd.PersistentFlags().IntVar(&vars.Fanout, "fanout", 50, "how many command should be executed in parallel")

	return powerCmd
}
