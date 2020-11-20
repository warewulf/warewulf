package poweroff

import (
	"github.com/spf13/cobra"
)

var (
	powerCmd = &cobra.Command{
		Use:   "poweroff",
		Short: "power off node(s)",
		Long:  "turn power off for one or more nodes",
		RunE:  CobraRunE,
	}
)

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return powerCmd
}
