package delete

import (
	"github.com/spf13/cobra"
	"github.com/warewulf/warewulf/internal/pkg/node"
)

var (
	baseCmd = &cobra.Command{
		DisableFlagsInUseLine: true,
		Use:                   "delete [OPTIONS] NODE [NODE ...]",
		Short:                 "Delete a node from Warewulf",
		Long:                  "This command will remove NODE(s) from the Warewulf node configuration.",
		Args:                  cobra.MinimumNArgs(1),
		RunE:                  CobraRunE,
		Aliases:               []string{"rm", "del", "remove"},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) != 0 {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}

			nodeDB, _ := node.New()
			nodes := nodeDB.ListAllNodes()
			return nodes, cobra.ShellCompDirectiveNoFileComp
		},
	}
	SetYes   bool
	SetForce bool // no hash checking, so always using force
)

func init() {
	SetForce = true
	baseCmd.PersistentFlags().BoolVarP(&SetYes, "yes", "y", false, "Set 'yes' to all questions asked")

}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}
