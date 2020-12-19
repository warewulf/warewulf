package powercycle

import (
	"github.com/spf13/cobra"
)

var (
	powerCmd = &cobra.Command{
		Use:   "powercycle",
		Short: "power cycle node(s)",
		Long:  "cycle power for one or more nodes",
		RunE:  CobraRunE,
	}
)

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return powerCmd
}
