package completions

import (
	"os"

	"github.com/spf13/cobra"
)

func CobraRunE(cmd *cobra.Command, args []string) (err error) {
	myArg := "bash"
	if len(args) == 1 {
		myArg = args[0]
	}
	switch myArg {
	case "zsh":
		err = cmd.Parent().Parent().GenZshCompletion(os.Stdout)
	case "fish":
		err = cmd.Parent().Parent().GenFishCompletion(os.Stdout, true)
	default:
		err = cmd.Parent().Parent().GenBashCompletion(os.Stdout)
	}
	return
}
