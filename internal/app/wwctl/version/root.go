package version

import (
	"github.com/spf13/cobra"
	"github.com/warewulf/warewulf/internal/app/wwctl/completions"
)

var (
	baseCmd = &cobra.Command{
		Use:               "version",
		Short:             "Version information",
		Long:              "This command will print the Warewulf version.",
		RunE:              CobraRunE,
		Args:              cobra.NoArgs,
		Aliases:           []string{"vers"},
		ValidArgsFunction: completions.None,
	}
	ListFull bool
)

func init() {
	baseCmd.PersistentFlags().BoolVarP(&ListFull, "full", "f", false, "List all compiled in variables.")
}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}
