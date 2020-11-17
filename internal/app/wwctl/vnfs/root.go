package vnfs

import (
	"fmt"
	"github.com/spf13/cobra"
)

var (
	baseCmd = &cobra.Command{
		Use:                "vnfs",
		Short:              "VNFS image management",
		Long:               "Virtual Node File System (VNFS) image management",
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
	fmt.Printf("Vnfs: Hello World\n")
	return nil
}