package image

import (
	"github.com/spf13/cobra"
	"github.com/warewulf/warewulf/internal/app/wwctl/image/build"
	"github.com/warewulf/warewulf/internal/app/wwctl/image/copy"
	"github.com/warewulf/warewulf/internal/app/wwctl/image/delete"
	"github.com/warewulf/warewulf/internal/app/wwctl/image/exec"
	"github.com/warewulf/warewulf/internal/app/wwctl/image/imprt"
	"github.com/warewulf/warewulf/internal/app/wwctl/image/kernels"
	"github.com/warewulf/warewulf/internal/app/wwctl/image/list"
	"github.com/warewulf/warewulf/internal/app/wwctl/image/rename"
	"github.com/warewulf/warewulf/internal/app/wwctl/image/shell"
	"github.com/warewulf/warewulf/internal/app/wwctl/image/show"
	"github.com/warewulf/warewulf/internal/app/wwctl/image/syncuser"
)

var baseCmd = &cobra.Command{
	DisableFlagsInUseLine: true,
	Use:                   "image COMMAND [OPTIONS]",
	Short:                 "Operating system image management",
	Long: "Starting with version 4, Warewulf uses images to build the bootable\n" +
		"node images. These commands will help you import, manage, and transform\n" +
		"images into bootable Warewulf images.",
	Aliases: []string{"vnfs", "container"},
}

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
	baseCmd.AddCommand(rename.GetCommand())
	baseCmd.AddCommand(kernels.GetCommand())
}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}
