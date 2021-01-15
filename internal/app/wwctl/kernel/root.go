package kernel

import (
	"github.com/hpcng/warewulf/internal/app/wwctl/kernel/imprt"
	"github.com/hpcng/warewulf/internal/app/wwctl/kernel/list"
	"github.com/spf13/cobra"
)

var (
	baseCmd = &cobra.Command{
		Use:   "kernel",
		Short: "Kernel Image Management",
		Long: "This command is for management of Warewulf Kernels to be used for\n" +
			"bootstrapping nodes",
	}
)

func init() {
	baseCmd.AddCommand(imprt.GetCommand())
	baseCmd.AddCommand(list.GetCommand())
}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}
