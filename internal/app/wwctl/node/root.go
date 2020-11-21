package node

import (
	"github.com/hpcng/warewulf/internal/app/wwctl/node/add"
	"github.com/hpcng/warewulf/internal/app/wwctl/node/list"
	"github.com/hpcng/warewulf/internal/app/wwctl/node/poweron"
	"github.com/hpcng/warewulf/internal/app/wwctl/node/poweroff"
	"github.com/hpcng/warewulf/internal/app/wwctl/node/powerstatus"
	"github.com/hpcng/warewulf/internal/app/wwctl/node/set"
	"github.com/spf13/cobra"
)

var (
	baseCmd = &cobra.Command{
		Use:                "node",
		Short:              "Node management",
		Long:               "Management of node settings and power management",
	}
)

func init() {
	baseCmd.AddCommand(poweron.GetCommand())
	baseCmd.AddCommand(poweroff.GetCommand())
	baseCmd.AddCommand(powerstatus.GetCommand())
	baseCmd.AddCommand(list.GetCommand())
	baseCmd.AddCommand(set.GetCommand())
	baseCmd.AddCommand(add.GetCommand())

}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}
