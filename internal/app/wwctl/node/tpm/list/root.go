package list

import (
	"github.com/spf13/cobra"
	"github.com/warewulf/warewulf/internal/app/wwctl/completions"
	"github.com/warewulf/warewulf/internal/pkg/hostlist"
)

var (
	keyFlag  bool
	jsonFlag bool
)

func GetCommand() *cobra.Command {
	baseCmd := &cobra.Command{
		DisableFlagsInUseLine: true,
		Use:                   "list [OPTIONS] [NODE NAME]",
		Short:                 "List TPM status of nodes",
		Long:                  "List TPM status of nodes in a tabular format.\n" + hostlist.Docstring,
		RunE:                  CobraRunE,
		ValidArgsFunction:     completions.Nodes,
		Args:                  cobra.ArbitraryArgs,
	}

	baseCmd.Flags().BoolVarP(&keyFlag, "key", "k", false, "Display the activation secret instead of the EK public key")
	baseCmd.Flags().BoolVarP(&jsonFlag, "json", "j", false, "Output data in JSON format")

	return baseCmd
}
