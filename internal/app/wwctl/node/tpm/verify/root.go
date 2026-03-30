package verify

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/warewulf/warewulf/internal/app/wwctl/completions"
	"github.com/warewulf/warewulf/internal/pkg/hostlist"
)

type variables struct {
	pcrFilter    []int
	displayEvent bool
}

func GetCommand() *cobra.Command {
	vars := variables{}
	baseCmd := &cobra.Command{
		DisableFlagsInUseLine: true,
		Use:                   "verify [OPTIONS] [NODE NAME]",
		Short:                 "Verify TPM quote for nodes",
		Long:                  "Verify TPM quote against the database for specified nodes.\n" + hostlist.Docstring,
		RunE:                  CobraRunE(&vars),
		ValidArgsFunction:     completions.Nodes,
		Args:                  cobra.ArbitraryArgs,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if cmd.Flags().Changed("pcr") && !vars.displayEvent {
				return fmt.Errorf("flag --pcr requires --eventlog")
			}
			return nil
		},
	}
	baseCmd.Flags().IntSliceVar(&vars.pcrFilter, "pcr", []int{}, "Optional filter for PCRs (comma separated)")
	baseCmd.Flags().BoolVar(&vars.displayEvent, "eventlog", false, "Display event log")
	return baseCmd
}
