package delete

import (
	"github.com/spf13/cobra"
	"github.com/warewulf/warewulf/internal/pkg/node"
)

type variables struct {
}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	vars := variables{}
	baseCmd := &cobra.Command{
		DisableFlagsInUseLine: true,
		Use:                   "list [RESOURCE]",
		Short:                 "list resource",
		Long:                  "This list all values the named RESOURCE. If no resource is named a list of the resources is printed.",
		Aliases:               []string{"show", "ls"},
		RunE:                  CobraRunE(&vars),
		Args:                  cobra.MaximumNArgs(1),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) != 0 {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			nodeDB, _ := node.New()
			return nodeDB.ListAllResources(), cobra.ShellCompDirectiveNoFileComp
		},
	}
	return baseCmd
}
