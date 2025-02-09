package kernels

import (
	"github.com/spf13/cobra"

	"github.com/warewulf/warewulf/internal/app/wwctl/completions"
)

var (
	baseCmd = &cobra.Command{
		DisableFlagsInUseLine: true,
		Use:                   "kernels [OPTIONS]",
		Short:                 "List available image kernels",
		Long:                  "This command lists the kernels that are available in the imported images.",
		RunE:                  CobraRunE,
		Aliases:               []string{"kernel"},
		ValidArgsFunction:     completions.Images,
	}
)

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}
