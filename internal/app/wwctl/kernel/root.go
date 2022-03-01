package kernel

import (
	"github.com/hpcng/warewulf/internal/app/wwctl/kernel/delete"
	"github.com/hpcng/warewulf/internal/app/wwctl/kernel/imprt"
	"github.com/hpcng/warewulf/internal/app/wwctl/kernel/list"
	"github.com/spf13/cobra"
)

var (
	baseCmd = &cobra.Command{
		DisableFlagsInUseLine: true,
		Use:                   "kernel COMMAND [OPTIONS]",
		Short:                 "Kernel Image Management",
		Long:                  "This command manages Warewulf Kernels used for bootstrapping nodes",
	}
)

func init() {
	baseCmd.AddCommand(imprt.GetCommand())
	baseCmd.AddCommand(list.GetCommand())
	baseCmd.AddCommand(delete.GetCommand())
}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}
