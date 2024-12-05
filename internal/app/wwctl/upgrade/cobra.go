package upgrade

import (
	"github.com/spf13/cobra"

	"github.com/warewulf/warewulf/internal/app/wwctl/upgrade/config"
	"github.com/warewulf/warewulf/internal/app/wwctl/upgrade/nodes"
)

var (
	Command = &cobra.Command{
		DisableFlagsInUseLine: true,
		Use:                   "upgrade <config|nodes> [OPTIONS]",
		Short:                 "Upgrade configuration files",
		Long: `Upgrade warewulf.conf or nodes.conf from a previous version of Warewulf 4 to a format
supported by the current version.`,
	}
)

func init() {
	Command.AddCommand(config.Command)
	Command.AddCommand(nodes.Command)
}
