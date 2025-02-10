package show

import (
	"log"

	"github.com/spf13/cobra"
	"github.com/warewulf/warewulf/internal/app/wwctl/completions"
	"github.com/warewulf/warewulf/internal/pkg/node"
)

var (
	baseCmd = &cobra.Command{
		DisableFlagsInUseLine: true,
		Use:                   "show [OPTIONS] OVERLAY_NAME FILE",
		Short:                 "Show (cat) a file within a Warewulf Overlay",
		Long:                  "This command displays the contents of FILE within OVERLAY_NAME.",
		RunE:                  CobraRunE,
		Aliases:               []string{"cat"},
		Args:                  cobra.ExactArgs(2),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) < 2 {
				return completions.OverlayAndFiles(cmd, args, toComplete)
			}
			return completions.None(cmd, args, toComplete)
		},
	}
	NodeName string
	Quiet    bool
)

func init() {
	baseCmd.PersistentFlags().StringVarP(&NodeName, "render", "r", "", "node used for the variables in the template")
	if err := baseCmd.RegisterFlagCompletionFunc("render", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		nodeDB, _ := node.New()
		return nodeDB.ListAllNodes(), cobra.ShellCompDirectiveNoFileComp
	}); err != nil {
		log.Println(err)
	}
	baseCmd.PersistentFlags().BoolVarP(&Quiet, "quiet", "q", false, "do not print information if multiple, backup files are written")
}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}
