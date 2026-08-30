package version

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/warewulf/warewulf/internal/pkg/version"
)

func CobraRunE(cmd *cobra.Command, args []string) error {

	fmt.Println("wwctl version:\t", version.Version())
	return nil
}
