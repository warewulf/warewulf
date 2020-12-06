package container

import (
	"github.com/hpcng/warewulf/internal/app/wwctl/container/build"
	"github.com/hpcng/warewulf/internal/app/wwctl/container/list"
	"github.com/hpcng/warewulf/internal/app/wwctl/container/pull"
	"github.com/spf13/cobra"
)

var (
	baseCmd = &cobra.Command{
		Use:   "container",
		Short: "VNFS image management",
		Long:  "Virtual Node File System (VNFS) image management",
	}
	test bool
)

func init() {
	baseCmd.AddCommand(build.GetCommand())
	baseCmd.AddCommand(list.GetCommand())
	baseCmd.AddCommand(pull.GetCommand())

}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}
