package version

import "github.com/spf13/cobra"

var (
	baseCmd = &cobra.Command{
		Use:     "version",
		Short:   "Version information",
		Long:    "This command will print the Warewulf version.",
		RunE:    CobraRunE,
		Args:    cobra.ExactArgs(0),
		Aliases: []string{"vers"},
	}
)

func init() {
}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}
