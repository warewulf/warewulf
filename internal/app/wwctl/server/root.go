package server

import (
	"github.com/hpcng/warewulf/internal/app/wwctl/server/start"
	"github.com/hpcng/warewulf/internal/app/wwctl/server/status"
	"github.com/hpcng/warewulf/internal/app/wwctl/server/stop"
	"github.com/spf13/cobra"
)

var (
	baseCmd = &cobra.Command{
		Use:   "server",
		Short: "Warewulf server process commands",
		Long:  "Warewulf profiles...",
	}
	test bool
)

func init() {
	baseCmd.AddCommand(start.GetCommand())
	baseCmd.AddCommand(status.GetCommand())
	baseCmd.AddCommand(stop.GetCommand())

}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}
