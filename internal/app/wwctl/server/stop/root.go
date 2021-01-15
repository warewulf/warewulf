package stop

import "github.com/spf13/cobra"

var (
	baseCmd = &cobra.Command{
		Use:   "stop",
		Short: "Stop Warewulf server",
		RunE:  CobraRunE,
	}
)

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}
