package clean

import (
	"github.com/spf13/cobra"
	"github.com/warewulf/warewulf/internal/app/wwctl/completions"
)

type variables struct {
}

func GetCommand() *cobra.Command {
	vars := variables{}
	baseCmd := &cobra.Command{
		DisableFlagsInUseLine: true,
		Use:                   "clean",
		Short:                 "Clean up",
		Long:                  "This command cleans the OCI cache and removes leftovers from deleted nodes",
		RunE:                  CobraRunE(&vars),
		Args:                  cobra.NoArgs,
		ValidArgsFunction:     completions.None,
	}
	return baseCmd
}
