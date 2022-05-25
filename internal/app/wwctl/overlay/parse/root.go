package parse

import (
	"log"

	"github.com/hpcng/warewulf/internal/pkg/node"
	"github.com/spf13/cobra"
)

var (
	baseCmd = &cobra.Command{
		DisableFlagsInUseLine: true,
		Use:                   "parse [OPTIONS] FILE_NAME",
		Short:                 "parses interprets the given file as warewulf template",
		Long:                  "This command tries to parse a given file as warewulf template, it needs not to be in overlay folder.",
		RunE:                  CobraRunE,
		Aliases:               []string{"parse"},
		Args:                  cobra.ExactArgs(1),
	}
	NodeName string
)

func init() {
	baseCmd.PersistentFlags().StringVarP(&NodeName, "node", "n", "", "node used for the variables in the template")
	if err := baseCmd.RegisterFlagCompletionFunc("node", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		nodeDB, _ := node.New()
		nodes, _ := nodeDB.FindAllNodes()
		var node_names []string
		for _, node := range nodes {
			node_names = append(node_names, node.Id.Get())
		}
		return node_names, cobra.ShellCompDirectiveNoFileComp
	}); err != nil {
		log.Println(err)
	}
}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}
