package container

import (
	"github.com/hpcng/warewulf/internal/app/wwctl/container/build"
	"github.com/hpcng/warewulf/internal/app/wwctl/container/delete"
	"github.com/hpcng/warewulf/internal/app/wwctl/container/exec"
	"github.com/hpcng/warewulf/internal/app/wwctl/container/imprt"
	"github.com/hpcng/warewulf/internal/app/wwctl/container/list"
	"github.com/spf13/cobra"
)

var (
	baseCmd = &cobra.Command{
		Use:   "container",
		Short: "Container / VNFS image management",
		Long: "Starting with version 4, Warewulf uses containers to build the bootable VNFS\n" +
			"images for nodes to boot. These commands will help you import, management, and\n" +
			"transform containers into bootable Warewulf VNFS images.",
	}
	test bool
)

func init() {
	baseCmd.AddCommand(build.GetCommand())
	baseCmd.AddCommand(list.GetCommand())
	baseCmd.AddCommand(imprt.GetCommand())
	baseCmd.AddCommand(exec.GetCommand())
	baseCmd.AddCommand(delete.GetCommand())
}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}
