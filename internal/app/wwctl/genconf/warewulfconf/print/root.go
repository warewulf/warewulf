package print

import (
	"github.com/spf13/cobra"
	"github.com/warewulf/warewulf/internal/app/wwctl/completions"
)

var (
	baseCmd = &cobra.Command{
		Use:               "print",
		Short:             "print wareweulf.conf",
		Long:              "This command prints the actual used warewulf.conf, can be used to create an empty valid warewulf.conf",
		RunE:              CobraRunE,
		Args:              cobra.NoArgs,
		ValidArgsFunction: completions.None,
	}
)

func GetCommand() *cobra.Command {
	return baseCmd
}
