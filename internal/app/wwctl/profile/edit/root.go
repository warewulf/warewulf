package edit

import (
	"github.com/spf13/cobra"
	"github.com/warewulf/warewulf/internal/app/wwctl/completions"
)

var (
	baseCmd = &cobra.Command{
		DisableFlagsInUseLine: true,
		Use:                   "edit [OPTIONS] PROFILE",
		Short:                 "Edit profile(s) with editor",
		Long:                  "This command opens an editor for the given profiles.",
		RunE:                  CobraRunE,
		ValidArgsFunction:     completions.Profiles,
		Args:                  cobra.ArbitraryArgs,
	}
	NoHeader bool
	Yes      bool
)

func init() {
	baseCmd.PersistentFlags().BoolVar(&NoHeader, "noheader", false, "Do not print header")
	baseCmd.PersistentFlags().BoolVarP(&Yes, "yes", "y", false, "Always confirm")
	_ = baseCmd.PersistentFlags().MarkHidden("yes")
}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}
