package group

import (
	"github.com/hpcng/warewulf/internal/app/wwctl/group/add"
	"github.com/hpcng/warewulf/internal/app/wwctl/group/delete"
	"github.com/hpcng/warewulf/internal/app/wwctl/group/list"
	"github.com/hpcng/warewulf/internal/app/wwctl/group/set"
	"github.com/spf13/cobra"
)

var (
	baseCmd = &cobra.Command{
		Use:                "group",
		Short:              "Group management",
		Long:               "Management of group settings and power management",
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
