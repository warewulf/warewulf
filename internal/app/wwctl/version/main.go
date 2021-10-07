package version

import (
	"fmt"

	"github.com/hpcng/warewulf/internal/pkg/version"
	"github.com/spf13/cobra"
)

func CobraRunE(cmd *cobra.Command, args []string) error {

	fmt.Println("Version:\t", version.GetVersion())

	return nil
}
