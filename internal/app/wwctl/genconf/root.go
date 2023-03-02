package genconf

import (
	"github.com/hpcng/warewulf/internal/app/wwctl/genconf/completions"
	"github.com/hpcng/warewulf/internal/app/wwctl/genconf/dfaults"
	"github.com/hpcng/warewulf/internal/app/wwctl/genconf/man"
	"github.com/hpcng/warewulf/internal/app/wwctl/genconf/reference"
	"github.com/hpcng/warewulf/internal/app/wwctl/genconf/warewulfconf"
	"github.com/hpcng/warewulf/internal/app/wwctl/genconf/man"
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
	ListFull  bool
	WWctlRoot *cobra.Command
)

func init() {
	baseCmd.AddCommand(completions.GetCommand())
	baseCmd.AddCommand(man.GetCommand())
	baseCmd.AddCommand(reference.GetCommand())
	baseCmd.AddCommand(dfaults.GetCommand())
	baseCmd.AddCommand(warewulfconf.GetCommand())
}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}
