package imprt

import (
	"github.com/spf13/cobra"
)

var (
	baseCmd = &cobra.Command{
		DisableFlagsInUseLine: true,
		Use:                   "import [OPTIONS] NODENAME",
		Short:                 "Import node(s) from yaml file",
		Long:                  "This command imports all the nodes defined in a file. It will overwrite nodes with same name.",
		RunE:                  CobraRunE,
		Args:                  cobra.MinimumNArgs(1),
		Aliases:               []string{"imprt"},
	}
)

func init() {
}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}
