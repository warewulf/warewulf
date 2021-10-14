package node

import (
	"github.com/hpcng/warewulf/internal/app/wwctl/node/add"
	"github.com/hpcng/warewulf/internal/app/wwctl/node/console"
	"github.com/hpcng/warewulf/internal/app/wwctl/node/delete"
	"github.com/hpcng/warewulf/internal/app/wwctl/node/list"
	"github.com/hpcng/warewulf/internal/app/wwctl/node/ready"
	"github.com/hpcng/warewulf/internal/app/wwctl/node/sensors"
	"github.com/hpcng/warewulf/internal/app/wwctl/node/set"
	"github.com/spf13/cobra"
)

var (
	baseCmd = &cobra.Command{
		DisableFlagsInUseLine: true,
		Use:   "node COMMAND [OPTONS]",
		Short: "Node management",
		Long:  "Management of node settings",
	}
)

func init() {
	baseCmd.AddCommand(sensors.GetCommand())
	baseCmd.AddCommand(list.GetCommand())
	baseCmd.AddCommand(set.GetCommand())
	baseCmd.AddCommand(add.GetCommand())
	baseCmd.AddCommand(delete.GetCommand())
	baseCmd.AddCommand(console.GetCommand())
	baseCmd.AddCommand(ready.GetCommand())
}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}
