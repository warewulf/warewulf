package list

import (
	"github.com/hpcng/warewulf/internal/pkg/node"
	"github.com/spf13/cobra"
)

var (
	baseCmd = &cobra.Command{
		DisableFlagsInUseLine: true,
		Use:                   "list [OPTIONS] [PATTERN]",
		Short:                 "List nodes",
		Long: "This command lists all configured nodes. Optionally, it will list only\n" +
			"nodes matching a glob PATTERN.",
		RunE:    CobraRunE,
		Aliases: []string{"ls"},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) != 0 {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}

			nodeDB, _ := node.ReadNodeYaml()
			nodes, _ := nodeDB.FindAllNodes()
			var node_names []string
			for _, node := range nodes {
				node_names = append(node_names, node.Id.Get())
			}
			return node_names, cobra.ShellCompDirectiveNoFileComp
		},
	}
	ShowNet  bool
	ShowIpmi bool
	ShowAll  bool
	ShowLong bool
)

func init() {
	baseCmd.PersistentFlags().BoolVarP(&ShowNet, "net", "n", false, "Show node network configurations")
	baseCmd.PersistentFlags().BoolVarP(&ShowIpmi, "ipmi", "i", false, "Show node IPMI configurations")
	baseCmd.PersistentFlags().BoolVarP(&ShowAll, "all", "a", false, "Show all node configurations")
	baseCmd.PersistentFlags().BoolVarP(&ShowLong, "long", "l", false, "Show long or wide format")
}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}
