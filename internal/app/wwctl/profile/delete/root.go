package delete

import (
	"github.com/spf13/cobra"
	"github.com/warewulf/warewulf/internal/app/wwctl/completions"
)

var (
	baseCmd = &cobra.Command{
		DisableFlagsInUseLine: true,
		Use:                   "delete [OPTIONS] PROFILE",
		Short:                 "Delete a node profile",
		Long:                  "This command deletes the node PROFILE. You may use a pattern for PROFILE.",
		Aliases:               []string{"remove", "rm", "del"},
		RunE:                  CobraRunE,
		Args:                  cobra.MinimumNArgs(1),
		ValidArgsFunction:     completions.Profiles,
	}
	SetYes bool
)

func init() {
	baseCmd.PersistentFlags().BoolVarP(&SetYes, "yes", "y", false, "Set 'yes' to all questions asked")
}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}
