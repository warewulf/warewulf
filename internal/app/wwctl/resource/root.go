package resource

import (
	"github.com/spf13/cobra"
	"github.com/warewulf/warewulf/internal/app/wwctl/resource/add"
	"github.com/warewulf/warewulf/internal/app/wwctl/resource/delete"
	"github.com/warewulf/warewulf/internal/app/wwctl/resource/list"
	"github.com/warewulf/warewulf/internal/app/wwctl/resource/set"
)

var (
	baseCmd = &cobra.Command{
		DisableFlagsInUseLine: true,
		Use:                   "resource COMMAND [OPTIONS]",
		Short:                 "manage resources",
		Long:                  "Management of global resources",
	}
)

func init() {
	baseCmd.AddCommand(list.GetCommand())
	baseCmd.AddCommand(set.GetCommand())
	baseCmd.AddCommand(add.GetCommand())
	baseCmd.AddCommand(delete.GetCommand())
}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}
