package rename

import (
	"github.com/spf13/cobra"
	"github.com/warewulf/warewulf/internal/app/wwctl/completions"
)

var baseCmd = &cobra.Command{
	DisableFlagsInUseLine: true,
	Use:                   "rename IMAGE NEW_NAME",
	Aliases:               []string{"mv"},
	Short:                 "Rename an existing image",
	Long:                  "This command will rename an existing image.",
	RunE:                  CobraRunE,
	Args:                  cobra.ExactArgs(2),
	ValidArgsFunction:     completions.Images(1),
}

var SetBuild bool

func init() {
	baseCmd.PersistentFlags().BoolVarP(&SetBuild, "build", "b", false, "Build image after rename")
}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}
