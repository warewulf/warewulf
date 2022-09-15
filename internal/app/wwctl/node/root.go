package node

import (
	"github.com/hpcng/warewulf/internal/app/wwctl/node/add"
	"github.com/hpcng/warewulf/internal/app/wwctl/node/console"
	"github.com/hpcng/warewulf/internal/app/wwctl/node/delete"
	"github.com/hpcng/warewulf/internal/app/wwctl/node/edit"
	"github.com/hpcng/warewulf/internal/app/wwctl/node/list"
	"github.com/hpcng/warewulf/internal/app/wwctl/node/sensors"
	"github.com/hpcng/warewulf/internal/app/wwctl/node/set"
	nodestatus "github.com/hpcng/warewulf/internal/app/wwctl/node/status"
	"github.com/spf13/cobra"
)

var (
	baseCmd = &cobra.Command{
		DisableFlagsInUseLine: true,
		Use:                   "node COMMAND [OPTONS]",
		Short:                 "Node management",
		Long: "Management of node settings. All node ranges can use brackets to identify\n" +
			"node ranges. For example: n00[00-4].cluster[0-1] will identify the first 5 nodes\n" +
			"in cluster0 and cluster1.",
	}
)

func init() {
	baseCmd.AddCommand(sensors.GetCommand())
	baseCmd.AddCommand(list.GetCommand())
	baseCmd.AddCommand(set.GetCommand())
	baseCmd.AddCommand(add.GetCommand())
	baseCmd.AddCommand(delete.GetCommand())
	baseCmd.AddCommand(console.GetCommand())
	baseCmd.AddCommand(nodestatus.GetCommand())
	baseCmd.AddCommand(edit.GetCommand())
}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}
