package reload

import "github.com/spf13/cobra"

var (
	baseCmd = &cobra.Command{
		Use:   "reload",
		Short: "Reload the Warewulf server configuration",
		RunE:  CobraRunE,
	}
)

func init() {
}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}
