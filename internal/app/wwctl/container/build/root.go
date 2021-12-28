package build

import (
	"github.com/hpcng/warewulf/internal/pkg/container"
	"github.com/spf13/cobra"
)

var (
	baseCmd = &cobra.Command{
		DisableFlagsInUseLine: true,
		Use:   "build [OPTIONS] CONTAINER [...]",
		Short: "(Re)build a bootable VNFS image",
		Long:  "This command will build a bootable VNFS image from imported CONTAINER image(s).",
		RunE:  CobraRunE,
		Args:  cobra.MinimumNArgs(1),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) != 0 {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			list, _ := container.ListSources()
			return list, cobra.ShellCompDirectiveNoFileComp
		},
	}
	BuildForce bool
	BuildAll   bool
	SetDefault bool
)

func init() {
	baseCmd.PersistentFlags().BoolVarP(&BuildAll, "all", "a", false, "(re)Build all VNFS images for all nodes")
	baseCmd.PersistentFlags().BoolVarP(&BuildForce, "force", "f", false, "Force rebuild, even if it isn't necessary")
	baseCmd.PersistentFlags().BoolVar(&SetDefault, "setdefault", false, "Set this container for the default profile")
}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}
