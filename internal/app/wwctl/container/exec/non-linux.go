// +build !linux

package exec

import (
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/spf13/cobra"
)

func CobraRunE(cmd *cobra.Command, args []string) error {
	wwlog.Printf(wwlog.ERROR, "This command does not work on non-Linux hosts\n")

	return nil
}
