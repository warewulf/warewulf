package profile

import (
	"github.com/spf13/cobra"
	"github.com/warewulf/warewulf/internal/app/wwctl/profile/add"
	"github.com/warewulf/warewulf/internal/app/wwctl/profile/delete"
	"github.com/warewulf/warewulf/internal/app/wwctl/profile/edit"
	"github.com/warewulf/warewulf/internal/app/wwctl/profile/list"
	"github.com/warewulf/warewulf/internal/app/wwctl/profile/set"
)

var (
	baseCmd = &cobra.Command{
		DisableFlagsInUseLine: true,
		Use:                   "profile COMMAND [OPTIONS]",
		Short:                 "Node configuration profile management",
		Long:                  "Management of node profile settings",
		Args:                  cobra.NoArgs,
	}
)

func init() {
	baseCmd.AddCommand(list.GetCommand())
	baseCmd.AddCommand(set.GetCommand())
	baseCmd.AddCommand(add.GetCommand())
	baseCmd.AddCommand(delete.GetCommand())
	baseCmd.AddCommand(edit.GetCommand())
}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}
