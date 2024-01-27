package warewulfconf

import (
	"github.com/spf13/cobra"
	"github.com/warewulf/warewulf/internal/app/wwctl/genconf/warewulfconf/print"
)

var (
	baseCmd = &cobra.Command{
		Use:     "warewulfconf",
		Short:   "access warewulf.conf",
		Long:    "Commands for accessing the actual used warewulf.conf",
		Args:    cobra.ExactArgs(0),
		Aliases: []string{"cnf"},
	}
	ListFull  bool
	WWctlRoot *cobra.Command
)

func init() {
	baseCmd.AddCommand(print.GetCommand())
}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}
