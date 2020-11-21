package set

import "github.com/spf13/cobra"

var (
	baseCmd = &cobra.Command{
		Use:                "set",
		Short:              "Set node configurations",
		Long:               "Set node configurations ",
		RunE:				CobraRunE,
	}
	SetVnfs string
)

func init() {
	baseCmd.PersistentFlags().StringVarP(&SetVnfs, "vnfs", "V", "", "Set node Virtual Node File System (VNFS)")

}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}
