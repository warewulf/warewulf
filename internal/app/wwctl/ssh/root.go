package ssh

import (
	"github.com/spf13/cobra"
)

var (
	baseCmd = &cobra.Command{
		DisableFlagsInUseLine: true,
		Use:                   "ssh [OPTIONS] NODE_PATTERN...",
		Short:                 "SSH into configured nodes in parallel",
		Long:                  "Easily ssh into nodes in parallel to run non-interactive commands\n",
		RunE:                  CobraRunE,
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
