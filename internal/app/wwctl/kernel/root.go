package kernel

import (
	"github.com/spf13/cobra"
	"github.com/warewulf/warewulf/internal/app/wwctl/kernel/list"
)

var (
	baseCmd = &cobra.Command{
		DisableFlagsInUseLine: true,
		Use:                   "kernel COMMAND [OPTIONS]",
		Short:                 "Kernel Management",
		Long:                  "This command manages kernels available to Warewulf",
	}
)

func init() {
	baseCmd.AddCommand(list.GetCommand())
}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}
