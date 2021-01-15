package status

import "github.com/spf13/cobra"

var (
	baseCmd = &cobra.Command{
		Use:   "status",
		Short: "Warewulf server status",
		RunE:  CobraRunE,
	}
)

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}
