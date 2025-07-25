package warewulfd

import (
	"github.com/spf13/cobra"
	"github.com/warewulf/warewulf/internal/app/wwctl/completions"
)

var (
	baseCmd = &cobra.Command{
		DisableFlagsInUseLine: true,
		Use:                   "warewulfd [OPTIONS]",
		Short:                 "Enable and start warewulfd",
		Long:                  "Enable and starts the warewulfd service.",
		RunE:                  CobraRunE,
		Args:                  cobra.NoArgs,
		ValidArgsFunction:     completions.None,
	}
)

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}
