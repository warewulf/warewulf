package off

import (
	"github.com/spf13/cobra"
	"github.com/warewulf/warewulf/internal/app/wwctl/completions"
)

type variables struct {
	Showcmd bool
	Fanout  int
}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	vars := variables{}
	powerCmd := &cobra.Command{
		DisableFlagsInUseLine: true,
		Use:                   "off [OPTIONS] [PATTERN ...]",
		Short:                 "Power off the given node(s)",
		Long:                  "This command will shutdown power to a set of nodes specified by PATTERN.",
		RunE:                  CobraRunE(&vars),
		Args:                  cobra.MinimumNArgs(1),
		ValidArgsFunction:     completions.Nodes,
	}
	powerCmd.PersistentFlags().BoolVarP(&vars.Showcmd, "show", "s", false, "only show command which will be executed")
	powerCmd.PersistentFlags().IntVar(&vars.Fanout, "fanout", 50, "how many command should be executed in parallel")

	return powerCmd
}
