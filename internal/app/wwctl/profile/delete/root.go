package delete

import (
	"github.com/spf13/cobra"
	"github.com/warewulf/warewulf/internal/pkg/node"
)

var (
	baseCmd = &cobra.Command{
		DisableFlagsInUseLine: true,
		Use:                   "delete [OPTIONS] PROFILE",
		Short:                 "Delete a node profile",
		Long:                  "This command deletes the node PROFILE. You may use a pattern for PROFILE.",
		Aliases:               []string{"remove", "rm", "del"},
		RunE:                  CobraRunE,
		Args:                  cobra.MinimumNArgs(1),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) != 0 {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			nodeDB, _ := node.New()
			return nodeDB.ListAllProfiles(), cobra.ShellCompDirectiveNoFileComp
		},
	}
	SetYes bool
)

func init() {
	baseCmd.PersistentFlags().BoolVarP(&SetYes, "yes", "y", false, "Set 'yes' to all questions asked")
}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}
