package vnfs

import (
	"github.com/hpcng/warewulf/internal/app/wwctl/vnfs/build"
	"github.com/hpcng/warewulf/internal/app/wwctl/vnfs/list"
	"github.com/spf13/cobra"
)

var (
	baseCmd = &cobra.Command{
		Use:                "vnfs",
		Short:              "VNFS image management",
		Long:               "Virtual Node File System (VNFS) image management",
	}
	test bool
)

func init() {
	baseCmd.AddCommand(build.GetCommand())
	baseCmd.AddCommand(list.GetCommand())
}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}

