package variables

import (
	"github.com/spf13/cobra"
	"github.com/warewulf/warewulf/internal/pkg/overlay"
)

var (
	baseCmd = &cobra.Command{
		Use:     "variables [flags] OVERLAY_NAME FILE_PATH",
		Short:   "Show variables for a template file in an overlay",
		Long:    "This command will show the variables for a given template file in a given\n" + "overlay.",
		Aliases: []string{"vars", "tags"},
		Args:    cobra.ExactArgs(2),
		RunE:    CobraRunE,
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) == 0 {
				return overlay.FindOverlays(), cobra.ShellCompDirectiveNoFileComp
			} else if len(args) == 1 {
				ov, err := overlay.Get(args[0])
				if err != nil {
					return nil, cobra.ShellCompDirectiveError
				}
				files, err := ov.GetFiles()
				if err != nil {
					return nil, cobra.ShellCompDirectiveError
				}
				return files, cobra.ShellCompDirectiveNoFileComp
			}
			return nil, cobra.ShellCompDirectiveNoFileComp
		},
	}
)

// GetCommand returns the root cobra.Command for this application.
func GetCommand() *cobra.Command {
	return baseCmd
}
