package child

import "github.com/spf13/cobra"

var (
	baseCmd = &cobra.Command{
		Use:                "__child",
		Hidden:             true,
		RunE:               CobraRunE,
		Args:               cobra.MinimumNArgs(1),
		FParseErrWhitelist: cobra.FParseErrWhitelist{UnknownFlags: true},
	}
	binds []string
)

func init() {
	baseCmd.PersistentFlags().StringArrayVarP(&binds, "bind", "b", []string{}, "bind points")
}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}
