package child

import "github.com/spf13/cobra"

var (
	baseCmd = &cobra.Command{
		Use:    "__child",
		Hidden: true,
		RunE:   CobraRunE,
		Args:   cobra.MinimumNArgs(1),
	}
)

func init() {

}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}
