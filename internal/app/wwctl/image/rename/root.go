package rename

import (
	"github.com/spf13/cobra"
	"github.com/warewulf/warewulf/internal/pkg/image"
)

var baseCmd = &cobra.Command{
	DisableFlagsInUseLine: true,
	Use:                   "rename IMAGE NEW_NAME",
	Aliases:               []string{"mv"},
	Short:                 "Rename an existing image",
	Long:                  "This command will rename an existing image.",
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

var SetBuild bool

func init() {
	baseCmd.PersistentFlags().BoolVarP(&SetBuild, "build", "b", false, "Build image after rename")
}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}
