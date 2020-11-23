package add

import "github.com/spf13/cobra"

var (
	baseCmd = &cobra.Command{
		Use:                "add",
		Short:              "Add a new node group",
		Long:               "Add a new node group ",
		RunE:				CobraRunE,
		Args: 				cobra.MinimumNArgs(1),
	}
	SetVnfs string
	SetKernel string
)

func init() {
	baseCmd.PersistentFlags().StringVarP(&SetVnfs, "vnfs", "V", "", "Set node Virtual Node File System (VNFS)")
	baseCmd.PersistentFlags().StringVarP(&SetKernel, "kernel", "K", "", "Set Kernel version for nodes")
}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}
