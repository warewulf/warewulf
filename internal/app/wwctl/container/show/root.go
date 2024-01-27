package show

import (
	"github.com/spf13/cobra"
	"github.com/warewulf/warewulf/internal/pkg/container"
)

var (
	baseCmd = &cobra.Command{
		DisableFlagsInUseLine: true,
		Use:                   "show [OPTIONS] CONTAINER",
		Short:                 "Show root fs dir for container",
		Long: `Shows the base directory for the chroot of the given container.
More information about the conainer can be shown with the '-a' option.`,
		RunE: CobraRunE,
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) != 0 {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			list, _ := container.ListSources()
			return list, cobra.ShellCompDirectiveNoFileComp
		},

		Args: cobra.MinimumNArgs(1),
	}
	ShowAll bool
)

func init() {
	baseCmd.PersistentFlags().BoolVarP(&ShowAll, "all", "a", false, "Show all information about a container")

}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}
