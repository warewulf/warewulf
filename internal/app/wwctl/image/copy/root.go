package copy

import (
	"github.com/spf13/cobra"
	"github.com/warewulf/warewulf/internal/pkg/image"
)

var (
	baseCmd = &cobra.Command{
		DisableFlagsInUseLine: true,
		Use:                   "copy IMAGE NEW_NAME",
		Aliases:               []string{"cp"},
		Short:                 "Copy an existing image",
		Long:                  "This command will duplicate an imported image.",
		RunE:                  CobraRunE,
		Args:                  cobra.MinimumNArgs(2),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) != 0 {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			list, _ := image.ListSources()
			return list, cobra.ShellCompDirectiveNoFileComp
		},
	}
	Build bool
)

func init() {
	baseCmd.PersistentFlags().BoolVarP(&Build, "build", "b", false, "Build image after copy")
}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}
