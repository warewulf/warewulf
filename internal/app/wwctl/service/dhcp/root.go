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
	ShowConfig bool
)

func init() {
	baseCmd.PersistentFlags().BoolVarP(&ShowConfig, "show", "s", false, "Show configuration rather than writing to files")

}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}
