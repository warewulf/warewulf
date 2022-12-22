package delete

import (
	"github.com/hpcng/warewulf/internal/pkg/container"
	"github.com/spf13/cobra"
)

var (
	baseCmd = &cobra.Command{
		DisableFlagsInUseLine: true,
		Use:                   "delete [OPTIONS] CONTAINER [...]",
		Short:                 "Delete an imported container",
		Long:                  "This command will delete CONTAINERs that have been imported into Warewulf.",
		RunE:                  CobraRunE,
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) != 0 {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			list, _ := container.ListSources()
			return list, cobra.ShellCompDirectiveNoFileComp
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
