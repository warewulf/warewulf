package set

import (
	"github.com/spf13/cobra"
	"github.com/warewulf/warewulf/internal/pkg/node"
)

type variables struct {
	tags map[string]string
}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	vars := variables{}
	baseCmd := &cobra.Command{
		DisableFlagsInUseLine: true,
		Use:                   "set RESOURCE",
		Short:                 "change a resource",
		Long:                  "This command changes the named RESOURCE. Use UNDEF to remove a tag.",
		Aliases:               []string{"modify", "change"},
		RunE:                  CobraRunE(&vars),
		Args:                  cobra.ExactArgs(1),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) != 0 {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			nodeDB, _ := node.New()
			return nodeDB.ListAllResources(), cobra.ShellCompDirectiveNoFileComp
		},
	}
	baseCmd.PersistentFlags().StringToStringVar(&vars.tags, "tagset", map[string]string{}, "set cluster wide resource tags")
	return baseCmd
}
