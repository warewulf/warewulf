package print

import (
	"github.com/spf13/cobra"
)

var (
	baseCmd = &cobra.Command{
		Use:   "print",
		Short: "print wareweulf.conf",
		Long:  "This command prints the actual used warewulf.conf, can be used to create an empty valid warewulf.conf",
		RunE:  CobraRunE,
		Args:  cobra.ExactArgs(0),
	}
)

func init() {
}

func GetCommand() *cobra.Command {
	return baseCmd
}
