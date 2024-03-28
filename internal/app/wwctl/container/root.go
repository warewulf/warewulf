package container

import (
	"github.com/spf13/cobra"
	"github.com/warewulf/warewulf/internal/app/wwctl/container/build"
	ocicache "github.com/warewulf/warewulf/internal/app/wwctl/container/cache"
	"github.com/warewulf/warewulf/internal/app/wwctl/container/copy"
	"github.com/warewulf/warewulf/internal/app/wwctl/container/delete"
	"github.com/warewulf/warewulf/internal/app/wwctl/container/exec"
	"github.com/warewulf/warewulf/internal/app/wwctl/container/imprt"
	"github.com/warewulf/warewulf/internal/app/wwctl/container/list"
	"github.com/warewulf/warewulf/internal/app/wwctl/container/reimprt"
	"github.com/warewulf/warewulf/internal/app/wwctl/container/rename"
	"github.com/warewulf/warewulf/internal/app/wwctl/container/shell"
	"github.com/warewulf/warewulf/internal/app/wwctl/container/show"
	"github.com/warewulf/warewulf/internal/app/wwctl/container/syncuser"
)

var baseCmd = &cobra.Command{
	DisableFlagsInUseLine: true,
	Use:                   "container COMMAND [OPTIONS]",
	Short:                 "Container / VNFS image management",
	Long: "Starting with version 4, Warewulf uses containers to build the bootable VNFS\n" +
		"node images. These commands will help you import, manage, and transform\n" +
		"containers into bootable Warewulf VNFS images.",
	Aliases: []string{"vnfs"},
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
	baseCmd.AddCommand(reimprt.GetCommand())
	baseCmd.AddCommand(rename.GetCommand())
	baseCmd.AddCommand(ocicache.GetCommand())
}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}
