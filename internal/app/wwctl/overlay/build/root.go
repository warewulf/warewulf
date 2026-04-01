package build

import (
	"github.com/spf13/cobra"
	"github.com/warewulf/warewulf/internal/app/wwctl/completions"
)

var (
	baseCmd = &cobra.Command{
		DisableFlagsInUseLine: true,
		Use:                   "build [OPTIONS] NODENAME...",
		Short:                 "(Re)build node overlays",
		Long:                  "This command builds overlays for given nodes.",
		RunE:                  CobraRunE,
		ValidArgsFunction:     completions.Nodes,
		Args:                  cobra.ArbitraryArgs,
	}
	Workers int
)

func init() {
	baseCmd.PersistentFlags().IntVar(&Workers, "workers", 0, "The number of parallel workers building overlays (<=0 indicates 1 worker per CPU)")
}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}
