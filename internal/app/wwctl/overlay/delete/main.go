package delete

import (
	"fmt"
	"github.com/spf13/cobra"
)

func CobraRunE(cmd *cobra.Command, args []string) error {
	fmt.Printf("Delete: Hello World\n")

	return nil
}