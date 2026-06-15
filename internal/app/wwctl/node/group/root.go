package group

import (
	"github.com/spf13/cobra"

	"github.com/warewulf/warewulf/internal/app/wwctl/node/group/list"
)

var baseCmd = &cobra.Command{
	DisableFlagsInUseLine: true,
	Use:                   "group COMMAND",
	Short:                 "Inspect nodegroups",
	Long:                  "Inspect nodegroups declared in nodes.conf or a node/profile nodegroups: field",
	Args:                  cobra.NoArgs,
}

func init() {
	baseCmd.AddCommand(list.GetCommand())
}

// GetCommand returns the `wwctl node group` subcommand tree.
func GetCommand() *cobra.Command {
	return baseCmd
}
