package list

import (
	"fmt"
	"github.com/spf13/cobra"
)

func CobraRunE(cmd *cobra.Command, args []string) error {
	fmt.Printf("This command is coming soon...\n")

	return nil
}