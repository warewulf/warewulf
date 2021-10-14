package server

import (
	"github.com/hpcng/warewulf/internal/app/wwctl/server/reload"
	"github.com/hpcng/warewulf/internal/app/wwctl/server/restart"
	"github.com/hpcng/warewulf/internal/app/wwctl/server/start"
	"github.com/hpcng/warewulf/internal/app/wwctl/server/status"
	"github.com/hpcng/warewulf/internal/app/wwctl/server/stop"
	"github.com/spf13/cobra"
)

var (
	baseCmd = &cobra.Command{
		DisableFlagsInUseLine: true,
		Use:   "server COMMAND [OPTIONS]",
		Short: "Warewulf server process commands",
		Long:  "This command will allow you to control the Warewulf daemon process.",
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
