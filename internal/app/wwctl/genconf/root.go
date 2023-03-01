package genconf

import (
	"github.com/hpcng/warewulf/internal/app/wwctl/genconf/completions"
	"github.com/spf13/cobra"
)

var (
	baseCmd = &cobra.Command{
		Use:     "genconfig",
		Short:   "Generate various configurations",
		Long:    "This command will allow you to generate different configurations like bash-completions.",
		Args:    cobra.ExactArgs(0),
		Aliases: []string{"cnf"},
	}
	ListFull bool
)

func init() {
	baseCmd.AddCommand(completions.GetCommand())
}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}
