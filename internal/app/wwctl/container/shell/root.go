package shell

import (
	"github.com/hpcng/warewulf/internal/pkg/container"
	"github.com/spf13/cobra"
)

var (
	baseCmd = &cobra.Command{
		DisableFlagsInUseLine: true,
		Use:                   "shell [OPTIONS] CONTAINER",
		Short:                 "Run a shell inside of a Warewulf container",
		Long:                  "Run a interactive shell inside of a warewulf CONTAINER.\n",
		RunE:                  CobraRunE,
		Args:                  cobra.MinimumNArgs(1),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) != 0 {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			list, _ := container.ListSources()
			return list, cobra.ShellCompDirectiveNoFileComp
		},
		FParseErrWhitelist: cobra.FParseErrWhitelist{UnknownFlags: true},
	}
	binds []string
)

func init() {
	baseCmd.PersistentFlags().StringArrayVarP(&binds, "bind", "b", []string{}, "Bind a local path into the container (must exist)")
}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}
