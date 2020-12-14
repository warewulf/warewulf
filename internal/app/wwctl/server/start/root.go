package start

import "github.com/spf13/cobra"

var (
	baseCmd = &cobra.Command{
		Use:   "start",
		Short: "Start Warewulf server",
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
