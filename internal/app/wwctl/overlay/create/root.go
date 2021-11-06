package create

import (
	"github.com/spf13/cobra"
)

var (
	baseCmd = &cobra.Command{
		Use:   "create [flags] (overlay kind) (overlay name)",
		Short: "Initialize a new Overlay",
		Long:  "This command will create a new empty overlay.",
		RunE:  CobraRunE,
		Args:  cobra.ExactArgs(2),
	}
)

func init() {
}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}
