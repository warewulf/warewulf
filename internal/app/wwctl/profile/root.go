package profile

import (
	"github.com/hpcng/warewulf/internal/app/wwctl/profile/add"
	"github.com/hpcng/warewulf/internal/app/wwctl/profile/delete"
	"github.com/hpcng/warewulf/internal/app/wwctl/profile/edit"
	"github.com/hpcng/warewulf/internal/app/wwctl/profile/list"
	"github.com/hpcng/warewulf/internal/app/wwctl/profile/set"
	"github.com/spf13/cobra"
)

var (
	baseCmd = &cobra.Command{
		DisableFlagsInUseLine: true,
		Use:                   "profile COMMAND [OPTIONS]",
		Short:                 "Node configuration profile management",
		Long:                  "Management of node profile settings",
		Aliases:               []string{"nodes"},
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
