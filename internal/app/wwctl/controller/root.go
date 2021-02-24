package controller

import (
	"github.com/hpcng/warewulf/internal/app/wwctl/controller/add"
	"github.com/hpcng/warewulf/internal/app/wwctl/controller/delete"
	"github.com/hpcng/warewulf/internal/app/wwctl/controller/list"
	"github.com/hpcng/warewulf/internal/app/wwctl/controller/set"
	"github.com/spf13/cobra"
)

var (
	baseCmd = &cobra.Command{
		Use:   "controller",
		Short: "Controller management",
		Long:  "Management of group settings and power management",
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
