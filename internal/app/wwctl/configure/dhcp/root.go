package dhcp

import (
	"github.com/spf13/cobra"
)

var (
	baseCmd = &cobra.Command{
		DisableFlagsInUseLine: true,
		Use:                   "dhcp [OPTIONS]",
		Short:                 "Manage and initialize DHCP",
		Long: "DHCP is a dependent service to Warewulf. This command will configure DHCP as defined\n" +
			"in the warewulf.conf file.",
		RunE: CobraRunE,
	}
	setShow bool
)

func init() {
	baseCmd.PersistentFlags().BoolVarP(&setShow, "show", "s", false, "Show configuration (don't update)")
}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}
