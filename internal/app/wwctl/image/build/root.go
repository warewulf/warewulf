package build

import (
	"github.com/spf13/cobra"
	"github.com/warewulf/warewulf/internal/pkg/image"
)

var (
	baseCmd = &cobra.Command{
		DisableFlagsInUseLine: true,
		Use:                   "build [OPTIONS] IMAGE [...]",
		Short:                 "(Re)build a bootable image",
		Long:                  "This command will build a bootable image from an imported IMAGE(s).",
		RunE:                  CobraRunE,
		Args:                  cobra.ArbitraryArgs,
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) != 0 {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			list, _ := image.ListSources()
			return list, cobra.ShellCompDirectiveNoFileComp
		},
	}
	BuildForce bool
	BuildAll   bool
	SyncUser   bool
)

func init() {
	baseCmd.PersistentFlags().BoolVarP(&BuildAll, "all", "a", false, "(re)Build all images")
	baseCmd.PersistentFlags().BoolVarP(&BuildForce, "force", "f", false, "Force rebuild, even if it isn't necessary")
	baseCmd.PersistentFlags().BoolVar(&SyncUser, "syncuser", false, "Synchronize UIDs/GIDs from host to image")
}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}
