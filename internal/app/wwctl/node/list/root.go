package list

import "github.com/spf13/cobra"

var (
	baseCmd = &cobra.Command{
		Use:     "list [flags] (node pattern)",
		Short:   "List nodes matching pattern",
		Long:    "This command will show you configured nodes.",
		RunE:    CobraRunE,
		Aliases: []string{"ls"},
	}
	ShowNet      bool
	ShowIpmi     bool
	ShowAll      bool
	ShowLong     bool
	ShowLastSeen bool
)

func init() {
	baseCmd.PersistentFlags().BoolVarP(&ShowNet, "net", "n", false, "Show node network configurations")
	baseCmd.PersistentFlags().BoolVarP(&ShowIpmi, "ipmi", "i", false, "Show node IPMI configurations")
	baseCmd.PersistentFlags().BoolVarP(&ShowAll, "all", "a", false, "Show all node configurations")
	baseCmd.PersistentFlags().BoolVarP(&ShowLong, "long", "l", false, "Show long or wide format")
	baseCmd.PersistentFlags().BoolVarP(&ShowLastSeen, "lastseen", "L", false, "Show last seen")
}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}
