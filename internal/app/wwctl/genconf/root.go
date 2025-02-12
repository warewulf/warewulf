package genconf

import (
	"github.com/spf13/cobra"
	"github.com/warewulf/warewulf/internal/app/wwctl/genconf/man"
	"github.com/warewulf/warewulf/internal/app/wwctl/genconf/reference"
	"github.com/warewulf/warewulf/internal/app/wwctl/genconf/warewulfconf"

	cobraCompletions "github.com/warewulf/warewulf/internal/app/wwctl/completions"
)

var (
	baseCmd = &cobra.Command{
		Use:               "genconfig",
		Short:             "Generate various configurations",
		Long:              "This command will allow you to generate different configurations like bash-completions.",
		Args:              cobra.NoArgs,
		Aliases:           []string{"cnf"},
		ValidArgsFunction: cobraCompletions.None,
		Hidden:            true,
	}
	ListFull  bool
	WWctlRoot *cobra.Command
)

func init() {
	baseCmd.AddCommand(man.GetCommand())
	baseCmd.AddCommand(reference.GetCommand())
	baseCmd.AddCommand(warewulfconf.GetCommand())
}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}
