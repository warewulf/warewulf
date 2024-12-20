package man

import (
	"github.com/spf13/cobra"
)

var (
	baseCmd = &cobra.Command{
		Use:     "man",
		Short:   "manpage generation",
		Long:    "This command generates the man pages for all commands, needs target dir as argument",
		RunE:    CobraRunE,
		Args:    cobra.ExactArgs(1),
		Aliases: []string{"man_pages"},
	}
)

func GetCommand() *cobra.Command {
	return baseCmd
}
