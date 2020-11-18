package copy

import (
	"fmt"
	"github.com/spf13/cobra"
)

func CobraRunE(cmd *cobra.Command, args []string) error {

	fmt.Printf("This will copy '%s' to overlay '%s'\n", args[1], args[0])

	return nil
}