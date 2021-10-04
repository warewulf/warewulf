package version

import (
	"fmt"

	"github.com/spf13/cobra"
)

var Version = "development"

func CobraRunE(cmd *cobra.Command, args []string) error {

	fmt.Println("Version foo:\t", Version)

	return nil
}
