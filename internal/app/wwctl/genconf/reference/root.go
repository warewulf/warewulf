package reference

import (
	"github.com/spf13/cobra"
)

var (
	baseCmd = &cobra.Command{
		Use:   "reference",
		Short: "reference generation",
		Long:  "This command generates the references in ReStructured Text, needs target dir as argument",
		RunE:  CobraRunE,
		Args:  cobra.ExactArgs(1),
	}
)

func GetCommand() *cobra.Command {
	return baseCmd
}
