package start

import "github.com/spf13/cobra"

var (
	baseCmd = &cobra.Command{
		DisableFlagsInUseLine: true,
		Use:                   "start [OPTIONS]",
		Short:                 "Start Warewulf server",
		RunE:                  CobraRunE,
	}
	SetForeground bool
)

func init() {
	baseCmd.PersistentFlags().BoolVarP(&SetForeground, "foreground", "f", false, "Run daemon process in the foreground")
}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}
