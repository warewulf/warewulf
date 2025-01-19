package syncuser

import (
	"github.com/spf13/cobra"
	"github.com/warewulf/warewulf/internal/pkg/image"
)

var (
	baseCmd = &cobra.Command{
		DisableFlagsInUseLine: true,
		Use:                   "syncuser [OPTIONS] IMAGE",
		Short:                 "Synchronizes user in image",
		Long: `Synchronize the uids and gids from the host to the image.
Users/groups which are only present in the image will be preserved if no
uid/gid collision is detected. File ownerships are also changed.`,
		RunE: CobraRunE,
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) != 0 {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			list, _ := image.ListSources()
			return list, cobra.ShellCompDirectiveNoFileComp
		},

		Args: cobra.MinimumNArgs(1),
	}
	write bool
	build bool
)

func init() {
	baseCmd.PersistentFlags().BoolVar(&write, "write", false, "Synchronize uis/gids and write files in image")
	baseCmd.PersistentFlags().BoolVar(&build, "build", false, "Build image after syncuser is completed")
}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}
