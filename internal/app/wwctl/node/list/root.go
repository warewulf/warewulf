package list

import "github.com/spf13/cobra"

var (
	baseCmd = &cobra.Command{
		DisableFlagsInUseLine: true,
		Use:     "list [OPTIONS] [PATTERN]",
		Short:   "List nodes",
		Long:    "This command lists all configured nodes. Optionally, it will list only\n" +
		         "nodes matching a glob PATTERN.",
		RunE:    CobraRunE,
		Aliases: []string{"ls"},
	}
	ShowNet  bool
	ShowIpmi bool
	ShowAll  bool
	ShowLong bool
)

func init() {
	baseCmd.PersistentFlags().BoolVarP(&ShowNet, "net", "n", false, "Show node network configurations")
	baseCmd.PersistentFlags().BoolVarP(&ShowIpmi, "ipmi", "i", false, "Show node IPMI configurations")
	baseCmd.PersistentFlags().BoolVarP(&ShowAll, "all", "a", false, "Show all node configurations")
	baseCmd.PersistentFlags().BoolVarP(&ShowLong, "long", "l", false, "Show long or wide format")
}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}
