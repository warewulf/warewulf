package container

import (
	"github.com/hpcng/warewulf/internal/app/wwctl/container/build"
	"github.com/hpcng/warewulf/internal/app/wwctl/container/copy"
	"github.com/hpcng/warewulf/internal/app/wwctl/container/delete"
	"github.com/hpcng/warewulf/internal/app/wwctl/container/exec"
	"github.com/hpcng/warewulf/internal/app/wwctl/container/imprt"
	"github.com/hpcng/warewulf/internal/app/wwctl/container/list"
	"github.com/hpcng/warewulf/internal/app/wwctl/container/shell"
	"github.com/hpcng/warewulf/internal/app/wwctl/container/show"
	"github.com/hpcng/warewulf/internal/app/wwctl/container/syncuser"
	"github.com/spf13/cobra"
)

var (
	baseCmd = &cobra.Command{
		DisableFlagsInUseLine: true,
		Use:                   "container COMMAND [OPTIONS]",
		Short:                 "Container / VNFS image management",
		Long: "Starting with version 4, Warewulf uses containers to build the bootable VNFS\n" +
			"node images. These commands will help you import, manage, and transform\n" +
			"containers into bootable Warewulf VNFS images.",
		Aliases: []string{"vnfs"},
	}
)

func init() {
	baseCmd.AddCommand(build.GetCommand())
	baseCmd.AddCommand(list.GetCommand())
	baseCmd.AddCommand(imprt.GetCommand())
	baseCmd.AddCommand(exec.GetCommand())
	baseCmd.AddCommand(shell.GetCommand())
	baseCmd.AddCommand(delete.GetCommand())
	baseCmd.AddCommand(show.GetCommand())
	baseCmd.AddCommand(syncuser.GetCommand())
	baseCmd.AddCommand(copy.GetCommand())

}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}
