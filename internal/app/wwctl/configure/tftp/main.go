package tftp

import (
	"github.com/hpcng/warewulf/internal/pkg/configure"
	"github.com/spf13/cobra"
)

func CobraRunE(cmd *cobra.Command, args []string) error {
	return configure.TFTP()
}
