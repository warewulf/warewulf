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
)

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}
