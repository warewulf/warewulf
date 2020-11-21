package delete

import "github.com/spf13/cobra"

var (
	baseCmd = &cobra.Command{
		Use:                "del",
		Short:              "Set node configurations",
		Long:               "Set node configurations ",
		RunE:				CobraRunE,
	}
	SetVnfs string
	SetKernel string
	//	SetGroupLevel bool
)

func init() {
	baseCmd.PersistentFlags().StringVarP(&SetVnfs, "vnfs", "V", "", "Set node Virtual Node File System (VNFS)")
	baseCmd.PersistentFlags().StringVarP(&SetKernel, "kernel", "K", "", "Set Kernel version for nodes")
}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}
