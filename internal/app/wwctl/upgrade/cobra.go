package upgrade

import (
	"github.com/spf13/cobra"

	"github.com/warewulf/warewulf/internal/app/wwctl/upgrade/config"
	"github.com/warewulf/warewulf/internal/app/wwctl/upgrade/nodes"
)

func GetCommand() *cobra.Command {
	command := &cobra.Command{
		DisableFlagsInUseLine: true,
		Use:                   "upgrade <config|nodes> [OPTIONS]",
		Short:                 "Upgrade configuration files",
		Long: `Upgrade warewulf.conf or nodes.conf from a previous version of Warewulf 4 to a format
supported by the current version.`,
	}
	command.AddCommand(config.GetCommand())
	command.AddCommand(nodes.GetCommand())
	return command
}
