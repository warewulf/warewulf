package copy

import (
	"github.com/spf13/cobra"
	"github.com/warewulf/warewulf/internal/pkg/container"
)

var (
	baseCmd = &cobra.Command{
		DisableFlagsInUseLine: true,
		Use:                   "copy CONTAINER NEW_NAME",
		Aliases:               []string{"cp"},
		Short:                 "Copy an existing container",
		Long:                  "This command will duplicate an imported container image.",
		RunE:                  CobraRunE,
		Args:                  cobra.MinimumNArgs(2),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) != 0 {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			list, _ := container.ListSources()
			return list, cobra.ShellCompDirectiveNoFileComp
		},
	}
)

func init() {
	// Nothing to do here
}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}
