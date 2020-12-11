package status

import "github.com/spf13/cobra"

var (
	baseCmd = &cobra.Command{
		Use:   "status",
		Short: "Warewulf server status",
		Long:  "Warewulf Server ",
		RunE:  CobraRunE,
	}
)

func init() {
}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}
