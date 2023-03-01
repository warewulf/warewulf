package completions

import "github.com/spf13/cobra"

var (
	baseCmd = &cobra.Command{
		Use:     "completions",
		Short:   "shell completion",
		Long:    "This command generates the bash completions if no argument is given.",
		RunE:    CobraRunE,
		Args:    cobra.MaximumNArgs(1),
		Aliases: []string{"bash"},
	}
	Zsh bool
)

func init() {
}

func GetCommand() *cobra.Command {
	return baseCmd
}
