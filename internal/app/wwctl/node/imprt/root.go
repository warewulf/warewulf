package imprt

import (
	"fmt"

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
		Aliases:               []string{"import"},
	}
	ImportCVS bool
)

func init() {
	baseCmd.PersistentFlags().BoolVarP(&ImportCVS, "cvs", "c", false, "Import CVS file")
}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}

func Usage() {
	fmt.Println(`The csv file must be structured in following way:
node,option1,option2,net.netname1.netopt
node01,value1,value2,net.netname1,netvalue`)
}
