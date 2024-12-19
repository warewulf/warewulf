package build

import (
	"log"
	"runtime"

	"github.com/spf13/cobra"
	"github.com/warewulf/warewulf/internal/pkg/node"
	"github.com/warewulf/warewulf/internal/pkg/overlay"
)

var (
	baseCmd = &cobra.Command{
		DisableFlagsInUseLine: true,
		Use:                   "build [OPTIONS] NODENAME...",
		Short:                 "(Re)build node overlays",
		Long:                  "This command builds overlays for given nodes.",
		RunE:                  CobraRunE,
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) != 0 {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}

			nodeDB, _ := node.New()
			nodes, _ := nodeDB.FindAllNodes()
			var node_names []string
			for _, node := range nodes {
				node_names = append(node_names, node.Id())
			}
			return node_names, cobra.ShellCompDirectiveNoFileComp
		},
	}
	OverlayNames []string
	OverlayDir   string
	Workers      int
)

func init() {
	baseCmd.PersistentFlags().StringSliceVarP(&OverlayNames, "overlay", "O", []string{}, "Build only specific overlay(s)")

	if err := baseCmd.RegisterFlagCompletionFunc("overlay", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		list, _ := overlay.FindOverlays()
		return list, cobra.ShellCompDirectiveNoFileComp
	}); err != nil {
		log.Println(err)
	}
	baseCmd.PersistentFlags().StringVarP(&OverlayDir, "output", "o", "", `Do not create an overlay image for distribution but write to
	the given directory. An overlay must also be ge given to use this option.`)
	baseCmd.PersistentFlags().IntVar(&Workers, "workers", runtime.NumCPU(), "The number of parallel workers building overlays")
}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}
