package child

import (
	"github.com/spf13/cobra"
)

var (
	baseCmd = &cobra.Command{
		DisableFlagsInUseLine: true,
		Use:                   "__child",
		Hidden:                true,
		RunE:                  CobraRunE,
		Args:                  cobra.MinimumNArgs(1),
		FParseErrWhitelist:    cobra.FParseErrWhitelist{UnknownFlags: true},
	}
	binds    []string
	nodename string
)

func init() {
	baseCmd.Flags().StringVarP(&nodename, "node", "n", "", "create ro overlay for given node")
	baseCmd.Flags().StringArrayVarP(&binds, "bind", "b", []string{}, "bind points")
}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}
