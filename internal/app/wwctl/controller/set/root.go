package set

import "github.com/spf13/cobra"

var (
	baseCmd = &cobra.Command{
		Use:   "set",
		Short: "Set",
		Long:  "Set",
		RunE:  CobraRunE,
		Args:  cobra.MinimumNArgs(1),
	}
	SetAll     bool
	SetIpaddr  string
	SetFqdn    string
	SetComment string
)

func init() {
	baseCmd.PersistentFlags().BoolVarP(&SetAll, "all", "a", false, "Set all controllers")
	baseCmd.PersistentFlags().StringVarP(&SetIpaddr, "ipaddr", "I", "", "Set the controller's IP address")
	baseCmd.PersistentFlags().StringVarP(&SetFqdn, "fqdn", "F", "", "Set the controller's FQDN")
	baseCmd.PersistentFlags().StringVarP(&SetComment, "comment", "C", "", "Comments describing this controller")

}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}
