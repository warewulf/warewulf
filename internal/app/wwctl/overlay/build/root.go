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
	BuildHost   bool
	BuildNodes  bool
	OverlayName string
	OverlayDir  string
)

func init() {
	baseCmd.PersistentFlags().BoolVarP(&BuildHost, "host", "H", false, "Build overlays only for the host")
	baseCmd.PersistentFlags().BoolVarP(&BuildNodes, "nodes", "N", false, "Build overlays only for the nodes")
	baseCmd.PersistentFlags().StringVarP(&OverlayName, "overlay", "O", "", "Build only specific overlay")
	baseCmd.PersistentFlags().StringVarP(&OverlayDir, "output", "o", "", `Do not create an overlay to image, for distribution but write to|
	the given directory. An overlay must also be ge given to use this option. '/dev/stdin' will print|
	the processed overlay.`)

}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}
