package syncuser

import (
	"github.com/spf13/cobra"
	"github.com/warewulf/warewulf/internal/pkg/image"
)

var (
	baseCmd = &cobra.Command{
		DisableFlagsInUseLine: true,
		Use:                   "syncuser [OPTIONS] IMAGE",
		Short:                 "Synchronize UIDs/GIDs from the server to an OS image",
		Long: `Synchronize UIDs and GIDs from the server into an OS image, updating
/etc/passwd, /etc/group, and file ownerships within the image. Users and
groups that exist only in the image are preserved unless a UID/GID collision
is detected.

This command affects the image itself (a one-time operation at build/import
time). To also make host users available on provisioned nodes at runtime, add
the "syncuser" overlay to the node or profile.`,
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
	baseCmd.PersistentFlags().BoolVar(&write, "write", true, "Synchronize UIDs/GIDs and write files in OS image")
	baseCmd.PersistentFlags().BoolVar(&build, "build", false, "Build OS image after syncuser is completed")
}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}
