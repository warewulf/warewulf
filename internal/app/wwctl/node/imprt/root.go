package imprt

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	baseCmd = &cobra.Command{
		DisableFlagsInUseLine: true,
		Use:                   "import [OPTIONS] FILE",
		Short:                 "Import node(s) from yaml FILE",
		Long:                  "This command imports all the nodes defined in a file. It will overwrite nodes with same name.",
		RunE:                  CobraRunE,
		Args:                  cobra.ExactArgs(1),
		Aliases:               []string{"import"},
	}
	ImportCSV bool
	setYes    bool
)

func init() {
	baseCmd.PersistentFlags().BoolVarP(&ImportCSV, "csv", "c", false, "Import CSV file")
	baseCmd.Flags().BoolVar(&ImportCSV, "cvs", false, "Import CSV file")
	baseCmd.Flags().Lookup("cvs").Hidden = true
	if err := baseCmd.Flags().MarkDeprecated("cvs", "use --csv instead"); err != nil {
		panic(err)
	}
	baseCmd.PersistentFlags().BoolVarP(&setYes, "yes", "y", false, "Set 'yes' to all questions asked")
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
