package power

import (
	"fmt"
	"github.com/spf13/cobra"
)

var (
	baseCmd = &cobra.Command{
		Use:                "power",
		Short:              "Node power management",
		Long:               "Node Power management commands",
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
	fmt.Printf("Power: Hello World\n")
	return nil
}