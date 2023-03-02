package dfaults

import "github.com/spf13/cobra"

var (
	baseCmd = &cobra.Command{
		Use:     "defaults",
		Short:   "print defaults",
		Long:    "This command prints the fallbacks which are used if defaults.conf isn't present",
		RunE:    CobraRunE,
		Args:    cobra.NoArgs,
		Aliases: []string{"dfaults"},
	}
	Zsh bool
)

func init() {
}

func GetCommand() *cobra.Command {
	return baseCmd
}
