package delete

import "github.com/spf13/cobra"

var (
	baseCmd = &cobra.Command{
		Use:   "delete [flags] [container name]...",
		Short: "Delete an imported container",
		Long:  "This command will delete a container that has been imported into Warewulf.",
		RunE:  CobraRunE,
		Args:  cobra.MinimumNArgs(1),
	}
)

func init() {

}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}
