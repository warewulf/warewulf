package show

import (
	"github.com/spf13/cobra"
)

var (
	baseCmd = &cobra.Command{
		Use:   "show [OPTIONS] {system|runtime} OVERLAY_NAME FILE",
		Short: "Show (cat) a file within a Warewulf Overlay",
		Long: "This command displays the contents of FILE within OVERLAY_NAME.",
		RunE:    CobraRunE,
		Aliases: []string{"cat"},
		Args:    cobra.ExactArgs(3),
	}
)

func init() {
}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}
