package server

import (
	"github.com/spf13/cobra"
	"github.com/warewulf/warewulf/internal/app/wwctl/server/reload"
	"github.com/warewulf/warewulf/internal/app/wwctl/server/restart"
	"github.com/warewulf/warewulf/internal/app/wwctl/server/start"
	"github.com/warewulf/warewulf/internal/app/wwctl/server/status"
	"github.com/warewulf/warewulf/internal/app/wwctl/server/stop"
)

var (
	baseCmd = &cobra.Command{
		DisableFlagsInUseLine: true,
		Use:                   "server COMMAND [OPTIONS]",
		Short:                 "Warewulf server process commands",
		Long:                  "This command will allow you to control the Warewulf daemon process.",
	}
)

func init() {
	baseCmd.AddCommand(start.GetCommand())
	baseCmd.AddCommand(status.GetCommand())
	baseCmd.AddCommand(stop.GetCommand())
	baseCmd.AddCommand(restart.GetCommand())
	baseCmd.AddCommand(reload.GetCommand())

}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}
