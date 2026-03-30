package tpm

import (
	"github.com/spf13/cobra"
	"github.com/warewulf/warewulf/internal/app/wwctl/node/tpm/check"
	"github.com/warewulf/warewulf/internal/app/wwctl/node/tpm/list"
	"github.com/warewulf/warewulf/internal/app/wwctl/node/tpm/verify"
)

func GetCommand() *cobra.Command {
	baseCmd := &cobra.Command{
		DisableFlagsInUseLine: true,
		Use:                   "tpm COMMAND [OPTIONS]",
		Short:                 "TPM management and verification",
		Long:                  "TPM management and verification",
	}
	baseCmd.AddCommand(list.GetCommand())
	baseCmd.AddCommand(verify.GetCommand())
	baseCmd.AddCommand(check.GetCommand())
	return baseCmd
}
