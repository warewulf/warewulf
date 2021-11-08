package list

import "github.com/spf13/cobra"

var (
	baseCmd = &cobra.Command{
		DisableFlagsInUseLine: true,
		Use:     "list [OPTIONS]",
		Short:   "List imported Warewulf containers",
		Long:    "This command will show you the containers that are imported into Warewulf.",
		RunE:    CobraRunE,
		Aliases: []string{"ls"},
	}
)

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}
