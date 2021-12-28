package list

import "github.com/spf13/cobra"

var (
	baseCmd = &cobra.Command{
		DisableFlagsInUseLine: true,
		Use:     "list [OPTIONS] [PROFILE ...]",
		Short:   "List profiles and configurations",
		Long:    "This command will display configurations for PROFILE.",
		RunE:    CobraRunE,
		Aliases: []string{"ls"},
	}
	ShowAll bool
)

func init() {
	baseCmd.PersistentFlags().BoolVarP(&ShowAll, "all", "a", false, "Show all node configurations")
}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}
