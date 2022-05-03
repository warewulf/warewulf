package syncuser

import (
	"github.com/hpcng/warewulf/internal/pkg/container"
	"github.com/spf13/cobra"
)

var (
	baseCmd = &cobra.Command{
		DisableFlagsInUseLine: true,
		Use:                   "syncuser [OPTIONS] CONTAINER",
		Short:                 "Synchronizes user in container",
		Long: `Synchronize the uids and gids from the host to the container.
Users/groups which are only present in the container will be preserved if no
uid/gid collision is detected. File ownerships are also changed.`,
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
	noSyncUser bool
)

func init() {
	baseCmd.PersistentFlags().BoolVar(&noSyncUser, "nosyncuser", true, "Don't synchronize uis/gods just check")
}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}
