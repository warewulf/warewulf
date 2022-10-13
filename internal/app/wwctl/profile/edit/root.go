package edit

import (
	"github.com/hpcng/warewulf/internal/pkg/node"
	"github.com/spf13/cobra"
)

var (
	baseCmd = &cobra.Command{
		DisableFlagsInUseLine: true,
		Use:                   "edit [OPTIONS] NODENAME",
		Short:                 "Edit node(s) with editor",
		Long:                  "This command opens an editor for the given profiles.",
		RunE:                  CobraRunE,
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) != 0 {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			nodeDB, _ := node.New()
			profiles, _ := nodeDB.FindAllProfiles()
			var p_names []string
			for _, profile := range profiles {
				p_names = append(p_names, profile.Id.Get())
			}
			return p_names, cobra.ShellCompDirectiveNoFileComp
		},
	}
	NoHeader bool
)

func init() {
	baseCmd.PersistentFlags().BoolVar(&NoHeader, "noheader", false, "Do not print header")
}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}
