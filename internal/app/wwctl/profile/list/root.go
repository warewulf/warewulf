package list

import (
	"github.com/spf13/cobra"
	"github.com/warewulf/warewulf/internal/pkg/node"
)

type variables struct {
	showAll  bool
	showYaml bool
	showJson bool
}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	vars := variables{}
	baseCmd := &cobra.Command{
		DisableFlagsInUseLine: true,
		Use:                   "list [OPTIONS] [PROFILE ...]",
		Short:                 "List profiles and configurations",
		Long:                  "This command will display configurations for PROFILE.",
		RunE:                  CobraRunE(&vars),
		Aliases:               []string{"ls"},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) != 0 {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			nodeDB, _ := node.New()
			return nodeDB.ListAllProfiles(), cobra.ShellCompDirectiveNoFileComp
		},
	}
	baseCmd.PersistentFlags().BoolVarP(&vars.showAll, "all", "a", false, "Show all profile configurations")
	baseCmd.PersistentFlags().BoolVarP(&vars.showYaml, "yaml", "y", false, "Show profile configurations via yaml format")
	baseCmd.PersistentFlags().BoolVarP(&vars.showJson, "json", "j", false, "Show profile configurations via json format")

	return baseCmd
}
