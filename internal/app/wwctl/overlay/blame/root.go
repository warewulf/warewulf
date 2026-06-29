package blame

import (
	"github.com/spf13/cobra"
	"github.com/warewulf/warewulf/internal/app/wwctl/completions"
)

type variables struct {
	PathPrefix      string
	ShowModeChanges bool
	Format          string
}

// GetCommand returns the cobra.Command for overlay blame.
func GetCommand() *cobra.Command {
	vars := variables{}
	baseCmd := &cobra.Command{
		DisableFlagsInUseLine: true,
		Use:                   "blame [OPTIONS] NODE",
		Short:                 "Show which overlays provide node files",
		Long:                  "This command traces node overlay files back to their source overlays.",
		RunE:                  CobraRunE(&vars),
		Args:                  cobra.ExactArgs(1),
		ValidArgsFunction:     completions.Nodes,
	}
	baseCmd.PersistentFlags().StringVar(&vars.PathPrefix, "path-prefix", "", "Only show deployed paths under this prefix")
	baseCmd.PersistentFlags().BoolVar(&vars.ShowModeChanges, "show-mode-changes", false, "Include directory paths as mode-relevant contributors")
	baseCmd.PersistentFlags().StringVar(&vars.Format, "format", "table", "Output format: table|json")

	return baseCmd
}
