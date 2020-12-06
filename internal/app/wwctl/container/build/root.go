package build

import "github.com/spf13/cobra"

var (
	baseCmd = &cobra.Command{
		Use:   "build (container location | node search pattern)",
		Short: "VNFS Image Build",
		Long:  "VNFS kernel images",
		RunE:  CobraRunE,
		Args:  cobra.RangeArgs(0, 1),
	}
	BuildForce bool
	BuildAll   bool
)

func init() {
	baseCmd.PersistentFlags().BoolVarP(&BuildAll, "all", "a", false, "(re)Build all VNFS images for all nodes")
	baseCmd.PersistentFlags().BoolVarP(&BuildForce, "force", "f", false, "Force rebuild, even if it isn't necessary")
}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}
