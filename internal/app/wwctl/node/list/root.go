package list

import (
	"github.com/spf13/cobra"
	"github.com/warewulf/warewulf/internal/pkg/node"
)

type variables struct {
	showNet  bool
	showIpmi bool
	showAll  bool
	showLong bool
	showYaml bool
	showJson bool
}

func GetCommand() *cobra.Command {
	vars := variables{}
	baseCmd := &cobra.Command{
		DisableFlagsInUseLine: true,
		Use:                   "list [OPTIONS] [PATTERN]",
		Short:                 "List nodes",
		Long: "This command lists all configured nodes. Optionally, it will list only\n" +
			"nodes matching a PATTERN.",
		RunE:    CobraRunE(&vars),
		Aliases: []string{"ls"},
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
	baseCmd.PersistentFlags().BoolVarP(&vars.showNet, "net", "n", false, "Show node network configurations")
	baseCmd.PersistentFlags().BoolVarP(&vars.showIpmi, "ipmi", "i", false, "Show node IPMI configurations")
	baseCmd.PersistentFlags().BoolVarP(&vars.showAll, "all", "a", false, "Show all node configurations")
	baseCmd.PersistentFlags().BoolVarP(&vars.showLong, "long", "l", false, "Show long or wide format")
	baseCmd.PersistentFlags().BoolVarP(&vars.showYaml, "yaml", "y", false, "Show yaml format")
	baseCmd.PersistentFlags().BoolVarP(&vars.showJson, "json", "j", false, "Show json format")

	return baseCmd
}
