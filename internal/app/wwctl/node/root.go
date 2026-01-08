package node

import (
	"github.com/spf13/cobra"
	"github.com/warewulf/warewulf/internal/app/wwctl/node/add"
	"github.com/warewulf/warewulf/internal/app/wwctl/node/console"
	"github.com/warewulf/warewulf/internal/app/wwctl/node/delete"
	"github.com/warewulf/warewulf/internal/app/wwctl/node/edit"
	"github.com/warewulf/warewulf/internal/app/wwctl/node/export"
	"github.com/warewulf/warewulf/internal/app/wwctl/node/imprt"
	"github.com/warewulf/warewulf/internal/app/wwctl/node/list"
	"github.com/warewulf/warewulf/internal/app/wwctl/node/sensors"
	"github.com/warewulf/warewulf/internal/app/wwctl/node/set"
	nodestatus "github.com/warewulf/warewulf/internal/app/wwctl/node/status"
	"github.com/warewulf/warewulf/internal/app/wwctl/node/unset"
)

var (
	baseCmd = &cobra.Command{
		DisableFlagsInUseLine: true,
		Use:                   "node COMMAND [OPTONS]",
		Short:                 "Node management",
		Long: "Management of node settings. All node ranges can use brackets to identify\n" +
			"node ranges. For example: n00[00-4].cluster[0-1] will identify the first 5 nodes\n" +
			"in cluster0 and cluster1.",
		Aliases: []string{"nodes"},
		Args:    cobra.NoArgs,
	}
)

func init() {
	baseCmd.AddCommand(sensors.GetCommand())
	baseCmd.AddCommand(list.GetCommand())
	baseCmd.AddCommand(set.GetCommand())
	baseCmd.AddCommand(unset.GetCommand())
	baseCmd.AddCommand(add.GetCommand())
	baseCmd.AddCommand(delete.GetCommand())
	baseCmd.AddCommand(console.GetCommand())
	baseCmd.AddCommand(nodestatus.GetCommand())
	baseCmd.AddCommand(edit.GetCommand())
	baseCmd.AddCommand(imprt.GetCommand())
	baseCmd.AddCommand(export.GetCommand())
}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}
