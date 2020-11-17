package node

import (
	"github.com/hpcng/warewulf/internal/app/wwctl/node/power"
	"github.com/spf13/cobra"
)

var (
	baseCmd = &cobra.Command{
		Use:                "node",
		Short:              "Node management",
		Long:               "Management of node settings and power management",
	}
	test bool
)

func init() {
	baseCmd.PersistentFlags().BoolVarP(&test, "test", "t", false, "Testing.")

	baseCmd.AddCommand(power.GetCommand())

}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}
