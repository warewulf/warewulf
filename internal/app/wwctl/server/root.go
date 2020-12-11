package server

import "github.com/spf13/cobra"

var (
	baseCmd = &cobra.Command{
		Use:   "server",
		Short: "Warewulf server process commands",
		Long:  "Warewulf profiles...",
	}
	test bool
)

func init() {
}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}
