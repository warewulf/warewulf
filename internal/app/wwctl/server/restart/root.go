package restart

import "github.com/spf13/cobra"

var (
	baseCmd = &cobra.Command{
		DisableFlagsInUseLine: true,
		Use:                   "restart [OPTIONS]",
		Short:                 "Restart the Warewulf server",
		RunE:                  CobraRunE,
	}
)

func init() {
}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}
