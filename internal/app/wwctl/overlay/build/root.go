package build

import "github.com/spf13/cobra"

var (
	baseCmd = &cobra.Command{
		DisableFlagsInUseLine: true,
		Use:                   "build [OPTIONS] NODENAME...",
		Short:                 "(Re)build node overlays",
		Long:                  "This command builds overlays for given nodes.",
		RunE:                  CobraRunE,
	}
	BuildHost  bool
	BuildNodes bool
)

func init() {
	baseCmd.PersistentFlags().BoolVarP(&BuildHost, "host", "H", false, "Build overlays only for the host")
	baseCmd.PersistentFlags().BoolVarP(&BuildNodes, "nodes", "N", false, "Build overlays only for the nodes")
}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}
