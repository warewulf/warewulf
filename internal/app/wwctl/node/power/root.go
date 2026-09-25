package power

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/warewulf/warewulf/internal/app/wwctl/completions"
	"github.com/warewulf/warewulf/internal/pkg/hostlist"
)

type variables struct {
	Showcmd bool
	Fanout  int
}

func addFlags(cmd *cobra.Command, vars *variables) {
	cmd.PersistentFlags().BoolVarP(&vars.Showcmd, "show", "s", false, "only show command which will be executed")
	cmd.PersistentFlags().IntVar(&vars.Fanout, "fanout", 50, "how many command should be executed in parallel")
}

func GetCommand() *cobra.Command {
	vars := variables{}
	var actionHelp strings.Builder
	for _, action := range Actions {
		fmt.Fprintf(&actionHelp, "  %-8s%s\n", action.Name, action.Short)
	}
	powerCmd := &cobra.Command{
		DisableFlagsInUseLine: true,
		Use:                   "power [OPTIONS] ACTION PATTERN ...",
		Short:                 "Control node power state",
		Long: "This command controls the power state of a set of nodes specified by PATTERN.\n\n" +
			"ACTION is one of:\n" + actionHelp.String() + "\n" + hostlist.Docstring,
		Args: cobra.MatchAll(cobra.MinimumNArgs(2), validateAction),
		RunE: CobraRunE(&vars),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) == 0 {
				return actionNames(), cobra.ShellCompDirectiveNoFileComp
			}
			return completions.Nodes(cmd, args, toComplete)
		},
	}
	addFlags(powerCmd, &vars)
	return powerCmd
}

func GetActionCommand(name string) *cobra.Command {
	action, ok := lookupAction(name)
	if !ok {
		panic(fmt.Sprintf("unknown power action: %s", name))
	}
	vars := variables{}
	runE := CobraRunE(&vars)
	powerCmd := &cobra.Command{
		DisableFlagsInUseLine: true,
		Use:                   action.Name + " [OPTIONS] PATTERN ...",
		Short:                 action.Short,
		Long:                  action.Long + "\n" + hostlist.Docstring,
		Args:                  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runE(cmd, append([]string{action.Name}, args...))
		},
		ValidArgsFunction: completions.Nodes,
	}
	addFlags(powerCmd, &vars)
	return powerCmd
}
