package version

import "github.com/spf13/cobra"

var (
	baseCmd = &cobra.Command{
		DisableFlagsInUseLine: true,
		Use:                   "version [OPTIONS]",
		Short:                 "Warewulf version",
		RunE:                  CobraRunE,
	}
)

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}
