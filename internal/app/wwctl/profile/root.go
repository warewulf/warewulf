package profile

import (
	"github.com/hpcng/warewulf/internal/app/wwctl/profile/add"
	"github.com/hpcng/warewulf/internal/app/wwctl/profile/delete"
	"github.com/hpcng/warewulf/internal/app/wwctl/profile/list"
	"github.com/hpcng/warewulf/internal/app/wwctl/profile/set"
	"github.com/spf13/cobra"
)

var (
	baseCmd = &cobra.Command{
		Use:                "profile",
		Short:              "Management of node configuration profiles",
		Long:               "Warewulf profiles...",
	}
	test bool
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
