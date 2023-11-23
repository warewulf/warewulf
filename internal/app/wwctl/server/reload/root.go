package reload

import "github.com/spf13/cobra"

var (
	baseCmd = &cobra.Command{
		DisableFlagsInUseLine: true,
		Use:                   "reload [OPTIONS]",
		Short:                 "Reload the Warewulf server configuration",
		RunE:                  CobraRunE,
	}
)

func init() {
}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}
