package overlay

import (
	"github.com/spf13/cobra"
	"github.com/warewulf/warewulf/internal/app/wwctl/overlay/build"
	"github.com/warewulf/warewulf/internal/app/wwctl/overlay/chmod"
	"github.com/warewulf/warewulf/internal/app/wwctl/overlay/chown"
	"github.com/warewulf/warewulf/internal/app/wwctl/overlay/create"
	"github.com/warewulf/warewulf/internal/app/wwctl/overlay/delete"
	"github.com/warewulf/warewulf/internal/app/wwctl/overlay/edit"
	"github.com/warewulf/warewulf/internal/app/wwctl/overlay/imprt"
	"github.com/warewulf/warewulf/internal/app/wwctl/overlay/info"
	"github.com/warewulf/warewulf/internal/app/wwctl/overlay/list"
	"github.com/warewulf/warewulf/internal/app/wwctl/overlay/mkdir"
	"github.com/warewulf/warewulf/internal/app/wwctl/overlay/show"
)

var (
	baseCmd = &cobra.Command{
		DisableFlagsInUseLine: true,
		Use:                   "overlay COMMAND [OPTIONS]",
		Short:                 "Warewulf Overlay Management",
		Long:                  "Management interface for Warewulf overlays",
		Args:                  cobra.NoArgs,
	}
)

func init() {
	baseCmd.AddCommand(list.GetCommand())
	baseCmd.AddCommand(show.GetCommand())
	baseCmd.AddCommand(create.GetCommand())
	baseCmd.AddCommand(edit.GetCommand())
	baseCmd.AddCommand(delete.GetCommand())
	baseCmd.AddCommand(mkdir.GetCommand())
	baseCmd.AddCommand(build.GetCommand())
	baseCmd.AddCommand(imprt.GetCommand())
	baseCmd.AddCommand(chmod.GetCommand())
	baseCmd.AddCommand(chown.GetCommand())
	baseCmd.AddCommand(info.GetCommand())
}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}
