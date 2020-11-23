package list

import "github.com/spf13/cobra"

var (
	baseCmd = &cobra.Command{
		Use:                "list",
		Short:              "List group configurations",
		Long:               "List group configurations ",
		RunE:				CobraRunE,
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
