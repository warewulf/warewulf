package reset

import (
	"github.com/spf13/cobra"
	"github.com/warewulf/warewulf/internal/app/wwctl/completions"
	"github.com/warewulf/warewulf/internal/pkg/hostlist"
)

func GetCommand() *cobra.Command {
	baseCmd := &cobra.Command{
		DisableFlagsInUseLine: true,
		Use:                   "reset [OPTIONS] [NODE NAME]",
		Short:                 "Reset TPM quote for nodes",
		Long:                  "Move the NEW TPM quote to Current for specified nodes, or remove Current if no NEW quote exists.\n" + hostlist.Docstring,
		RunE:                  CobraRunE,
		ValidArgsFunction:     completions.Nodes,
		Args:                  cobra.ArbitraryArgs,
	}
	return baseCmd
}
