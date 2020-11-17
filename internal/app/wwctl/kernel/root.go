package kernel

import (
	"fmt"
	"github.com/spf13/cobra"
)

var (
	baseCmd = &cobra.Command{
		Use:                "kernel",
		Short:              "Kernel Image Management",
		Long:               "Management of Warewulf Kernels to be used for bootstrapping nodes",
		RunE:				CobraRunE,
	}
	test bool
)

func init() {
	baseCmd.PersistentFlags().BoolVarP(&test, "test", "t", false, "Testing.")

}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}

func CobraRunE(cmd *cobra.Command, args []string) error {
	fmt.Printf("Kernel: Hello World\n")
	return nil
}