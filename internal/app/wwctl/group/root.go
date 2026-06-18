package group

import (
	"github.com/spf13/cobra"

	"github.com/warewulf/warewulf/internal/app/wwctl/group/list"
)

var baseCmd = &cobra.Command{
	DisableFlagsInUseLine: true,
	Use:                   "group COMMAND",
	Short:                 "Inspect node groups",
	Long:                  "Inspect groups declared on a node or profile groups: field",
	Args:                  cobra.NoArgs,
}

func init() {
	baseCmd.AddCommand(list.GetCommand())
}

// GetCommand returns the `wwctl group` subcommand tree.
func GetCommand() *cobra.Command {
	return baseCmd
}
