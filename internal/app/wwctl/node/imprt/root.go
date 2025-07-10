package imprt

import (
	"github.com/spf13/cobra"
)

var (
	baseCmd = &cobra.Command{
		DisableFlagsInUseLine: true,
		Use:                   "import [OPTIONS] FILE",
		Short:                 "Import node(s) from yaml FILE",
		Long:                  "This command imports all the nodes defined in a YAML file. It will overwrite nodes with same name.",
		RunE:                  CobraRunE,
		Args:                  cobra.ExactArgs(1),
		Aliases:               []string{"import"},
	}
	setYes bool
)

func init() {
	baseCmd.PersistentFlags().BoolVarP(&setYes, "yes", "y", false, "Set 'yes' to all questions asked")
}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}
