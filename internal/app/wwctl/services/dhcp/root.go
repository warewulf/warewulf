package dhcp

import (
	"github.com/spf13/cobra"
)

var (
	baseCmd = &cobra.Command{
		Use:   "dhcp",
		Short: "DHCP configuration",
		Long:  "DHCP Config",
		RunE:  CobraRunE,
	}
	test bool
)

func init() {
}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}
