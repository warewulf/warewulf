package kernel

import (
	"github.com/hpcng/warewulf/internal/app/wwctl/kernel/build"
	"github.com/hpcng/warewulf/internal/app/wwctl/kernel/export"
	"github.com/hpcng/warewulf/internal/app/wwctl/kernel/imprt"
	"github.com/hpcng/warewulf/internal/app/wwctl/kernel/list"
	"github.com/spf13/cobra"
)

var (
	baseCmd = &cobra.Command{
		Use:                "kernel",
		Short:              "Kernel Image Management",
		Long:               "Management of Warewulf Kernels to be used for bootstrapping nodes",
	}
)

func init() {
	baseCmd.AddCommand(build.GetCommand())
	baseCmd.AddCommand(export.GetCommand())
	baseCmd.AddCommand(imprt.GetCommand())
	baseCmd.AddCommand(list.GetCommand())
}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}