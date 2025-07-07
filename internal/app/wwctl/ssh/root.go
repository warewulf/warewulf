package ssh

import (
	"github.com/spf13/cobra"
	"github.com/warewulf/warewulf/internal/app/wwctl/completions"
	"github.com/warewulf/warewulf/internal/pkg/hostlist"
)

var (
	baseCmd = &cobra.Command{
		DisableFlagsInUseLine: true,
		Use:                   "ssh [OPTIONS] NODE_PATTERN COMMAND",
		Short:                 "SSH into configured nodes in parallel",
		Long:                  "Easily ssh into nodes in parallel to run non-interactive commands\n" + hostlist.Docstring,
		RunE:                  CobraRunE,
		Args:                  cobra.MinimumNArgs(2),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) == 0 {
				return completions.Nodes(cmd, args, toComplete)
			}
			return completions.None(cmd, args, toComplete)
		},
	}
	DryRun  bool
	FanOut  int
	Sleep   int
	SshPath string
)

func init() {
	baseCmd.PersistentFlags().BoolVarP(&DryRun, "dryrun", "n", false, "Show commands to run")
	baseCmd.PersistentFlags().IntVarP(&FanOut, "fanout", "f", 32, "How many connections to run in parallel")
	baseCmd.PersistentFlags().IntVarP(&Sleep, "sleep", "s", 0, "Seconds to sleep inbetween processes")
	baseCmd.PersistentFlags().StringVar(&SshPath, "rsh", "/usr/bin/ssh", "Path to use for RSH/SSH command")
}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}
