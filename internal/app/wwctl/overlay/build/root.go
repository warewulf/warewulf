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
)

func init() {
	//baseCmd.PersistentFlags().BoolVarP(&BuildAll, "all", "a", false, "Build overlays for all nodes")
}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}
