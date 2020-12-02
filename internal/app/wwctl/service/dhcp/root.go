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
	DoConfig bool
)

func init() {
	baseCmd.PersistentFlags().BoolVarP(&DoConfig, "configure", "c", false, "Do the DHCP Configuration")

}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}
