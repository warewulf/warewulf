package completions

import (
	"os"

	"github.com/spf13/cobra"
)

func CobraRunE(cmd *cobra.Command, args []string) error {
	myArg := "bash"
	if len(args) == 1 {
		myArg = args[0]
	}
	switch myArg {
	case "zsh":
		cmd.GenZshCompletion(os.Stdout)
	case "fish":
		cmd.GenFishCompletion(os.Stdout, true)
	default:
		cmd.GenBashCompletion(os.Stdout)
	}
	return nil
}
